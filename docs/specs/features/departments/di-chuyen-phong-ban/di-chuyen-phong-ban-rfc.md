# Đề Xuất Kỹ Thuật (RFC) — Di Chuyển Phòng Ban (Move Department)

## Mục Tiêu
Thiết kế giải pháp kỹ thuật, thuật toán chống chu trình, kiểm soát độ sâu, và chiến lược xử lý giới hạn DynamoDB Transaction khi di chuyển cây con (`UC-02`).

---

# 1. Bối Cảnh (Context)
- Đáp ứng Use Case `UC-02` và PRD `di-chuyen-phong-ban-prd.md`.
- Đây là bài toán khó nhất của Mini HRM do DynamoDB giới hạn tối đa 100 mục trong một lệnh `TransactWriteItems`.

---

# 2. Vấn Đề Kỹ Thuật Cần Giải Quyết
1. **Phát hiện chu trình (Cycle Detection)** siêu nhanh mà không cần duyệt đệ quy toàn bộ CSDL.
2. **Kiểm tra trần độ sâu 5 cấp**: Tính toán xem nhánh con dài nhất của phòng nguồn khi ghép vào cha mới có bị chạm trần hay không.
3. **Cập nhật tiền tố Path**: Thay thế `oldPrefix` bằng `newPrefix` trên hàng loạt bản ghi.
4. **Vượt trần DynamoDB Transaction**: Khi cây con có nhiều hơn số lượng mục mà một transaction cho phép.

---

# 3. Kiến Trúc Thuật Toán & Xử Lý

### 3.1. Thuật toán phát hiện chu trình O(1) dựa vào chuỗi Path
Do trường `path` lưu giữ toàn bộ tổ tiên từ gốc: `parent.path = "/root/deptA/deptB/"`.
```go
func IsDescendant(sourceDept, targetParent *Department) bool {
    if sourceDept.ID == targetParent.ID {
        return true // Chuyển vào chính nó
    }
    // Kiểm tra xem ID của phòng nguồn có nằm trong chuỗi path của cha mới không
    matchSegment := fmt.Sprintf("/%s/", sourceDept.ID)
    return strings.Contains(targetParent.Path, matchSegment)
}
```
-> **Độ phức tạp O(1)**: Không cần truy vấn đệ quy các cấp con! Nếu `targetParent.Path` chứa ID của phòng nguồn, điều đó chứng minh chắc chắn cha mới là con/cháu của phòng nguồn.

### 3.2. Thuật toán kiểm tra độ sâu tối đa
1. Lấy toàn bộ danh sách phòng con của `sourceDept` bằng truy vấn tiền tố GSI:
   `begins_with(GSI1SK, "PATH#" + sourceDept.Path)`.
2. Tìm phòng con có số lượng dấu gạch chéo `/` nhiều nhất để xác định chiều cao cây con: `subTreeHeight`.
3. Độ sâu mới: `newDepth = depth(targetParent) + 1 + subTreeHeight`.
4. Nếu `newDepth > 5` -> Hủy thao tác ngay từ đầu.

---

# 4. Quyết Định Thiết Kế: Xử Lý Trần DynamoDB Transaction

> [!CAUTION]
> **Bước 4 là chỗ để nghĩ, không phải chỗ để gõ.**
> DynamoDB giới hạn tối đa **100 mục trong một lệnh `TransactWriteItems`**.

### Quyết định của Mini HRM:
1. **Trường hợp cây con <= 50 phòng ban**:
   - Sử dụng **1 lệnh `TransactWriteItems` duy nhất** cập nhật toàn bộ phòng con + phòng nguồn + bản ghi `AuditLog`.
   - Đảm bảo tính nguyên tử tuyệt đối (All-or-Nothing).
2. **Trường hợp cây con lớn (> 50 phòng ban)**:
   - Áp dụng cơ chế **Khóa Trạng Thái (Two-Phase Execution)**:
     - **Bước 1**: Transaction cập nhật phòng nguồn sang `status = "MOVING"`, chặn mọi thao tác ghi khác vào cây này.
     - **Bước 2**: Chia danh sách phòng con thành các lô nhỏ (lô 25 items) và thực hiện `BatchWriteItem` cập nhật `path`.
     - **Bước 3**: Cập nhật phòng nguồn về lại `status = "ACTIVE"`, cập nhật `parentId` mới và ghi `AuditLog`.
     - **Cơ chế Rollback**: Nếu có bất kỳ lô nào thất bại ở Bước 2, hệ thống dùng snapshot ban đầu để phục hồi lại `path` cũ và chuyển phòng nguồn về `ACTIVE`.

---

# 5. Thiết Kế API
- **Endpoint**: `POST /api/v1/departments/{id}/move`
- **Headers**:
  - `Idempotency-Key: <UUIDv4>`
- **Request Body**:
  ```json
  {
    "newParentId": "dept_tech_block"
  }
  ```
- **Response (`HTTP 200 OK`)**:
  ```json
  {
    "success": true,
    "data": {
      "id": "dept_backend",
      "oldPath": "/root/old_block/dept_backend/",
      "newPath": "/root/dept_tech_block/dept_backend/",
      "affectedChildrenCount": 12
    }
  }
  ```
