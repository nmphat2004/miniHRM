# Phân Tách Công Việc (Tasks) — Tạo Phòng Ban Mới

## Mục Tiêu
Danh sách các đầu việc kỹ thuật thực thi được để hoàn thiện tính năng Tạo phòng ban (`UC-01`).

---

# 1. Danh Sách Tasks Chi Tiết

### Task 1: Thiết kế Model DynamoDB & Hàm Ghi Transaction
- **Context**: Triển khai tầng Repository trong Go.
- **Scope IN**: 
  - Tạo struct `Department` và `AuditLog`.
  - Viết hàm `CreateDepartmentWithTransaction(ctx, dept, auditLog, idempotencyKey)` sử dụng AWS SDK v2 DynamoDB `TransactWriteItems`.
  - Khóa partition `DEPT_CODE#<code>` để đảm bảo tính duy nhất.
- **Files**: `backend/internal/department/repository.go`, `backend/internal/department/model.go`.
- **Test Cases**: 
  - Ghi thành công 1 phòng ban mới.
  - Ghi đè mã đã có -> Ném lỗi `ErrDuplicateCode`.

### Task 2: Triển khai Service Nghiệp Vụ & Thuật Toán Tính Path
- **Context**: Tầng Service trong Go.
- **Scope IN**:
  - Validate `BR-PB-01`, `BR-PB-02`, `BR-PB-03`.
  - Tính toán chuỗi `path` chuẩn xác.
  - Kiểm tra trạng thái phòng cha phải là `ACTIVE`.
- **Files**: `backend/internal/department/service.go`.
- **Test Cases**:
  - Unit test thuật toán tính `path` cấp 1 đến 5.
  - Test lỗi khi vượt quá 5 cấp.

### Task 3: Xây Dựng HTTP Handler & Middleware Phân Quyền
- **Context**: Tầng API trong Go.
- **Scope IN**:
  - Endpoint `POST /api/v1/departments`.
  - Bắt buộc Header `Idempotency-Key`.
  - Middleware kiểm tra vai trò Quản trị viên (Admin).
  - Trả envelope chuẩn `{ "success": true, "data": ... }`.
- **Files**: `backend/internal/department/handler.go`, `backend/cmd/server/main.go`.

### Task 4: Xây Dựng Giao Diện Modal Tạo Phòng Ban (Next.js)
- **Context**: Frontend Next.js + shadcn/ui.
- **Scope IN**:
  - Form tạo phòng ban với validation Zod (mã, tên, cha).
  - Xử lý đủ 6 trạng thái giao diện UI.
  - Tự động sinh `Idempotency-Key` (UUIDv4) ở mỗi phiên mở modal.
- **Files**: `frontend/src/app/(app)/departments/_components/CreateDepartmentModal.tsx`.

---

# 2. Trạng Thái & Ước Tính Độ Phức Tạp
- **Độ phức tạp**: Trung bình (Medium).
- **Trạng thái**: TODO.
