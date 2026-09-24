# Phân Tách Công Việc (Tasks) — Di Chuyển Phòng Ban

## Mục Tiêu
Danh sách đầu việc chi tiết để hoàn thành Use Case khó nhất hệ thống (`UC-02`).

---

# 1. Danh Sách Tasks Chi Tiết

### Task 1: Xây Dựng Thuật Toán Cycle Detection & Chiều Cao Cây
- **Context**: Package `department` trong Go.
- **Scope IN**:
  - Hàm `CheckCycle(source, targetParent)` kiểm tra `targetParent.Path`.
  - Hàm `CalculateMaxSubtreeDepth(ctx, sourceDeptId)`.
- **Files**: `backend/internal/department/tree_helper.go`.
- **Test Cases**: 
  - Unit test chu trình với cây mẫu 5 nút.
  - Unit test chặn vượt cấp.

### Task 2: Cơ Chế Cập Nhật Path Hàng Loạt & Xử Lý DynamoDB Transaction
- **Context**: Repository DynamoDB.
- **Scope IN**:
  - Thuật toán tính toán `newPrefix` và cập nhật mảng các phòng con.
  - Phân nhánh: Nếu `len(items) <= 50` dùng `TransactWriteItems`; nếu lớn hơn dùng cơ chế khóa `MOVING` và rollback.
- **Files**: `backend/internal/department/repository_move.go`.
- **Test Cases**:
  - Di chuyển cây con 12 phòng ban -> Xác thực toàn bộ 12 path.
  - Giả lập lỗi ở item thứ 10 -> Chứng minh transaction rollback sạch sẽ.

### Task 3: API Endpoint & Ghi Nhận Audit Log
- **Context**: Controller Go.
- **Scope IN**:
  - Endpoint `POST /api/v1/departments/{id}/move`.
  - Ghi `AuditLog` (`action = "department.moved"`).
- **Files**: `backend/internal/department/handler.go`.

### Task 4: Giao Diện Kéo Thả / Modal Di Chuyển Cây Phòng Ban
- **Context**: Next.js App Router + shadcn/ui.
- **Scope IN**:
  - Hộp thoại chọn cha mới với cảnh báo preview trực quan.
  - Xử lý đủ 6 trạng thái UI (đặc biệt là trạng thái Đang lưu và Xung đột OCC).
- **Files**: `frontend/src/app/(app)/departments/_components/MoveDepartmentDialog.tsx`.

---

# 2. Trạng Thái & Ước Tính Độ Phức Tạp
- **Độ phức tạp**: Rất cao (Hard).
- **Trạng thái**: TODO.
