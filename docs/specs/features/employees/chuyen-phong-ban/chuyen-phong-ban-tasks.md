# Phân Tách Công Việc (Tasks) — Chuyển Phòng Ban Cho Nhân Viên

## 1. Danh Sách Tasks Chi Tiết

### Task 1: Repository Transaction Chuyển Phòng
- **Context**: Package `employee` trong Go.
- **Scope IN**: Viết hàm `TransferEmployee(ctx, empId, newDeptId, removeOldManagerRole)` bọc trong `TransactWriteItems`.
- **Files**: `backend/internal/employee/repository.go`.

### Task 2: Service Validate & API Handler
- **Context**: Tầng Service và Handler Go.
- **Scope IN**: Kiểm tra phòng mới `ACTIVE`, kiểm tra quyền Admin, ghi nhận AuditLog.
- **Files**: `backend/internal/employee/service.go`, `backend/internal/employee/handler.go`.

### Task 3: UI Modal Điều Chuyển Nhân Viên (Next.js)
- **Context**: Frontend Next.js + shadcn/ui.
- **Scope IN**: Hộp thoại chọn phòng đích, xử lý bước xác nhận gỡ Trưởng phòng, 6 trạng thái UI.
- **Files**: `frontend/src/app/(app)/employees/_components/TransferEmployeeDialog.tsx`.

---

# 2. Độ Phức Tạp & Trạng Thái
- **Độ phức tạp**: Trung bình (Medium).
- **Trạng thái**: TODO.
