# Phân Tách Công Việc (Tasks) — Nhật Ký & Khóa Lạc Quan

## 1. Danh Sách Tasks Chi Tiết

### Task 1: DynamoDB TransactWriteItems Helper Với AuditLog
- **Context**: Package `audit` trong Go.
- **Scope IN**: Struct `AuditLog`, helper thực thi transaction nguyên tử nghiệp vụ + audit.
- **Files**: `backend/internal/audit/repository.go`.

### Task 2: Idempotency Middleware Trong Go
- **Context**: Tầng Middleware Go API.
- **Scope IN**: Đọc header `Idempotency-Key`, lưu trữ và tra cứu kết quả kèm TTL 24h trên DynamoDB.
- **Files**: `backend/internal/middleware/idempotency.go`.

### Task 3: Hook Quản Lý 6 Trạng Thái UI & Idempotency Key (Next.js)
- **Context**: Frontend Next.js.
- **Scope IN**: Hook `useUIState()` và hook tự động sinh `Idempotency-Key` trước khi gửi request POST.
- **Files**: `frontend/src/hooks/useUIState.ts`, `frontend/src/lib/api-client.ts`.

---

# 2. Độ Phức Tạp & Trạng Thái
- **Độ phức tạp**: Trung bình (Medium).
- **Trạng thái**: TODO.
