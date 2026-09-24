# Phân Tách Công Việc (Tasks) — Đăng Nhập & Kẹp Phạm Vi

## 1. Danh Sách Tasks Chi Tiết

### Task 1: Cognito Adapter & Mock AuthService
- **Context**: Package `auth` trong Go.
- **Scope IN**: Interface `AuthService` và 2 bản cài đặt: `CognitoService` (AWS SDK) và `MockAuthService` (chạy local test).
- **Files**: `backend/internal/auth/service.go`, `backend/internal/auth/mock.go`.

### Task 2: Go Scoping Middleware & TTL Cache
- **Context**: Middleware Go API.
- **Scope IN**: Lấy thông tin phòng ban hiện hành của user từ CSDL với bộ đệm TTL 60s; gán `UserScope` vào `context.Context`.
- **Files**: `backend/internal/middleware/scoping.go`.

### Task 3: Kẹp Điều Kiện Truy Vấn & Xử Lý Lỗi 404
- **Context**: Repository Go.
- **Scope IN**: Bắt buộc các hàm `Get`, `List` nhận `UserScope` và nhúng vào câu query DynamoDB; trả 404 khi ngoài scope.
- **Files**: `backend/internal/employee/repository_scoped.go`.

### Task 4: Next.js BFF Route Handlers & Cookie Management
- **Context**: Frontend Next.js.
- **Scope IN**: Các route `/api/bff/auth/login`, `/api/bff/auth/logout`, thiết lập cookie `HttpOnly`.
- **Files**: `frontend/src/app/api/bff/auth/login/route.ts`.

---

# 2. Độ Phức Tạp & Trạng Thái
- **Độ phức tạp**: Cao (Hard).
- **Trạng thái**: TODO.
