# Bộ Công Nghệ Chính Thức & Kiến Trúc Vận Hành — Mini HRM

## 1. Triết Lý & Mục Tiêu

Mini HRM là hệ thống quản trị cơ cấu phòng ban và nhân sự cho doanh nghiệp quy mô 200–500 nhân sự. Hệ thống được thiết kế theo mô hình **Cloud-Native AWS Serverless** nhưng bắt buộc phải tuân thủ tiêu chuẩn:
> **Chạy hoàn toàn trên máy cục bộ — Không cần kết nối internet hay tài khoản AWS thật.**

Toàn bộ mức A của bài toán được đóng gói sao cho người chấm bài hoặc lập trình viên mới chỉ cần gõ duy nhất một lệnh:
```bash
docker compose up -d
make dev
# -> Mở trình duyệt tại http://localhost:3000
```

---

## 2. Bảng Công Nghệ Chốt

| Tầng kiến trúc | Công nghệ sử dụng | Phiên bản / Thư viện chuẩn | Vai trò & Lý do lựa chọn |
| :--- | :--- | :--- | :--- |
| **Giao diện Web (Frontend)** | Next.js (App Router) | Next.js 15+, TypeScript, React 19 | Server-side rendering kết hợp Client components tối ưu hiệu năng và SEO nội bộ |
| **Giao diện & Thành phần UI** | Tailwind CSS + shadcn/ui | Radix UI primitives, Lucide Icons | Thiết kế hiện đại, tinh gọn, kiểm soát toàn bộ giao diện 6 trạng thái |
| **Quản lý phiên Web** | Cookie `HttpOnly` | SameSite=Strict, Secure | Token xác thực **tuyệt đối không để xuống JavaScript trình duyệt** để chống rò rỉ XSS |
| **Máy chủ API (Backend)** | Go (Golang) | Go 1.23+ (chuẩn `net/http` hoặc `chi`) | Tốc độ cao, biên dịch nhị phân siêu nhẹ, khởi động lạnh dưới 10ms trên AWS Lambda |
| **Bộ định tuyến kép (Dual-Runtime)** | `aws-lambda-go-api-proxy` | Adapter chuẩn | **Cùng một Router Go** chạy được cả ở chế độ HTTP Server (Local) và Lambda Handler (Cloud) |
| **Cơ sở dữ liệu (Database)** | Amazon DynamoDB | Single-Table Design + GSI | NoSQL tốc độ mili-giây, scale linh hoạt, hỗ trợ ACID qua `TransactWriteItems` |
| **Môi trường DB cục bộ** | DynamoDB Local | Docker image `amazon/dynamodb-local` | Giả lập 100% API DynamoDB trên cổng 8000 |
| **Dịch vụ Định danh (Auth)** | AWS Cognito | User Pool + App Client | Quản lý danh tính, JWT token, phân quyền qua Groups/Roles |
| **Môi trường Auth cục bộ** | Cognito Local / Mock Adapter | Docker (`cognito-local` hoặc In-memory) | Cho phép đăng nhập, cấp token giả lập mà không cần tài khoản AWS |
| **Kiểm thử (Testing)** | Go Test (`testing`, `testify`) | Unit test, Mock DynamoDB/Cognito | Kiểm thử 100% các quy tắc nghiệp vụ `BR-*` độc lập không phụ thuộc hạ tầng mạng |

---

## 3. Kiến Trúc Thiết Kế Hệ Thống & Luồng Dữ Liệu

```
                                  ┌────────────────────────┐
                                  │   AWS Cognito User Pool│
                                  │   (Mật khẩu & Định danh)
                                  └───────────▲────────────┘
                                              │ Xác thực token
┌────────────────┐     Cookie     ┌───────────┴────────────┐     Single Table     ┌────────────────────────┐
│  Trình duyệt   │  HttpOnly JWT  │       Next.js BFF      │      JSON REST       │      Go API Server     │
│ (shadcn/ui +   │◄──────────────►│    (App Router SSR/    │◄────────────────────►│  (Quyền & Nghiệp vụ    │
│  Tailwind CSS) │                │     Quản lý phiên)     │                      │   Kẹp phạm vi Scoping) │
└────────────────┘                └────────────────────────┘                      └───────────┬────────────┘
                                                                                              │ TransactWrite / Query
                                                                                              ▼
                                                                                  ┌────────────────────────┐
                                                                                  │   Amazon DynamoDB      │
                                                                                  │   1 Bảng duy nhất      │
                                                                                  │   + Ít nhất 1 GSI      │
                                                                                  └────────────────────────┘
```

---

## 4. Các Quyết Định Kiến Trúc Cốt Lõi

### 4.1. Cognito — Tách Cổng Ra Khỏi Bản Cài (Interface Segregation)
Để tầng nghiệp vụ của Go không bị phụ thuộc cứng vào AWS SDK và có thể chạy kiểm thử tốc độ cao khi mất mạng:
- Khai báo một interface xác thực trong Go:
  ```go
  type AuthService interface {
      Authenticate(ctx context.Context, username, password string) (*AuthSession, error)
      GetUser(ctx context.Context, token string) (*UserInfo, error)
  }
  ```
- Triển khai 2 bản cài đặt:
  1. `CognitoAuthService`: Gọi trực tiếp đến AWS Cognito SDK (dùng trên môi trường Production/Staging).
  2. `MockAuthService`: Dùng cho môi trường Test và Local Development. Sử dụng map bộ nhớ hoặc Docker `cognito-local` để kiểm tra mật khẩu tức thì mà không cần mạng.

### 4.2. Go API — Một Bộ Định Tuyến Chạy Ở Hai Nơi
Bộ định tuyến HTTP (ví dụ sử dụng `chi.Router` hoặc `gin`) được xây dựng thuần túy theo chuẩn `http.Handler`.
- **Khi chạy trên máy cục bộ (Local Dev)**: Khởi chạy HTTP server chuẩn lắng nghe trên cổng `8080`:
  ```go
  http.ListenAndServe(":8080", router)
  ```
- **Khi triển khai lên AWS Lambda**: Bọc router bằng thư viện chuyển đổi `httpadapter`:
  ```go
  adapter := httpadapter.New(router)
  lambda.Start(adapter.ProxyWithContext)
  ```
Nhờ cơ chế này, toàn bộ logic định tuyến, middleware phân quyền, kẹp phạm vi được kiểm thử và vận hành đồng nhất 100% giữa môi trường dev và production.

### 4.3. Quản Lý Phiên Cookie HttpOnly
- JavaScript trên trình duyệt **tuyệt đối không được phép đọc token**.
- Khi người dùng đăng nhập thành công, Next.js BFF (Backend For Frontend) thiết lập cookie `session_token` với thuộc tính:
  - `HttpOnly = true` (chống tấn công XSS đánh cắp token).
  - `Secure = true` (chỉ gửi qua HTTPS).
  - `SameSite = Strict` (chống tấn công CSRF).
- Mọi cuộc gọi từ client lên Next.js tự động đính kèm cookie. Next.js BFF giải mã hoặc chuyển tiếp token an toàn sang Go API.

---

## 5. Cấu Trúc Mã Nguồn Dự Kiến (Monorepo Mini HRM)

```text
miniHRM/
├── docker-compose.yml           # Khởi chạy DynamoDB Local + Cognito Local
├── Makefile                     # make dev, make test, make build
├── docs/                        # Toàn bộ tài liệu đặc tả nghiệp vụ & kiến trúc
│   ├── README.md
│   ├── tech-stack.md
│   ├── tong-quan-nghiep-vu.md
│   └── specs/
│       ├── phong-ban-spec.md
│       ├── nhan-vien-spec.md
│       ├── phan-quyen-va-kep-pham-vi-spec.md
│       └── nhat-ky-va-tinh-huong-bien-spec.md
├── backend/                     # Mã nguồn Go (API & Domain Logic)
│   ├── cmd/
│   │   ├── server/main.go       # Entrypoint chạy Local HTTP
│   │   └── lambda/main.go       # Entrypoint chạy AWS Lambda
│   ├── internal/
│   │   ├── auth/                # Cognito interface, JWT parser, mock
│   │   ├── department/          # Nghiệp vụ phòng ban, thuật toán path
│   │   ├── employee/            # Nghiệp vụ nhân sự, kẹp phạm vi
│   │   ├── audit/               # Sổ nhật ký, atomic transaction
│   │   ├── repository/dynamo/   # DynamoDB Single-Table Adapter
│   │   └── middleware/          # Scoping middleware, OCC, Idempotency
│   └── go.mod
└── frontend/                    # Mã nguồn Next.js (App Router)
    ├── src/
    │   ├── app/                 # Các màn hình phân cấp, SSR, BFF route handlers
    │   ├── components/ui/       # Thư viện component shadcn/ui
    │   ├── hooks/               # Custom hooks quản lý 6 trạng thái UI
    │   └── lib/                 # Tiện ích gọi API, xử lý lỗi chuẩn
    ├── package.json
    └── tailwind.config.ts
```
