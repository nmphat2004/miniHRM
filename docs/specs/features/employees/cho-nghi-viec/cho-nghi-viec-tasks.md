# Phân Tách Công Việc (Tasks) — Cho Nhân Viên Nghỉ Việc

## 1. Danh Sách Tasks Chi Tiết

### Task 1: Repository Kiểm Tra Manager & Soft Delete
- **Context**: Go backend repository.
- **Scope IN**: Kiểm tra `managerId` của các phòng ban, bọc cập nhật status `RESIGNED` trong `TransactWriteItems`.
- **Files**: `backend/internal/employee/repository_resign.go`.

### Task 2: Service Logic & API Handler
- **Context**: Tầng Service và Controller.
- **Scope IN**: Validate ngày nghỉ việc, middleware quyền Admin, trả lỗi chuẩn khi vi phạm `BR-NV-04`.
- **Files**: `backend/internal/employee/service.go`, `backend/internal/employee/handler.go`.

### Task 3: UI Thao Tác Cho Nghỉ Việc & Bộ Lọc Danh Sách (Next.js)
- **Context**: Frontend Next.js.
- **Scope IN**: Modal xác nhận thôi việc, thêm filter checkbox *"Bao gồm nhân viên đã nghỉ việc"*, hiển thị badge `RESIGNED` màu xám.
- **Files**: `frontend/src/app/(app)/employees/_components/ResignDialog.tsx`.

---

# 2. Độ Phức Tạp & Trạng Thái
- **Độ phức tạp**: Thấp - Trung bình.
- **Trạng thái**: TODO.
