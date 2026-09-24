# Đề Xuất Kỹ Thuật (RFC) — Tính Năng Tạo Phòng Ban Mới

## Mục Tiêu
Thiết kế kiến trúc kỹ thuật, schema dữ liệu DynamoDB, hợp đồng API và giao dịch nguyên tử cho tính năng Tạo phòng ban mới (`UC-01`).

---

# 1. Bối Cảnh (Context)
- Triển khai theo yêu cầu của PRD `tao-phong-ban-prd.md` và Use Case `UC-01`.
- Hệ thống chạy trên kiến trúc Go + DynamoDB Single-Table Design.

---

# 2. Vấn Đề Kỹ Thuật Cần Giải Quyết
1. Kiểm tra tính duy nhất của mã `code` toàn công ty trong DynamoDB (NoSQL không có ràng buộc `UNIQUE` tự động như SQL).
2. Kiểm tra tính duy nhất của tên `name` trong cùng phòng cha.
3. Tính toán trường dẫn xuất `path` không sai lệch dấu phân cách `/`.
4. Ghi nguyên tử cả phòng ban, bản ghi bảo vệ mã duy nhất, và bản ghi nhật ký `AuditLog` trong cùng 1 lệnh `TransactWriteItems`.

---

# 3. Mục Tiêu Kỹ Thuật (Goals)
- Đảm bảo tính toàn vẹn dữ liệu tuyệt đối (ACID) qua DynamoDB Transaction.
- Hỗ trợ Idempotency chống tạo trùng lặp.
- Độ trễ phản hồi API < 100ms.

---

# 4. Không Nằm Trong Phạm Vi (Non-Goals)
- Không xử lý di chuyển phòng ban (thuộc RFC riêng `di-chuyen-phong-ban-rfc.md`).

---

# 5. Kiến Trúc Đề Xuất (Proposed Architecture)

```
   POST /api/v1/departments (Header: Idempotency-Key)
               │
               ▼
   ┌───────────────────────┐
   │  Idempotency Check    │ ─── Trùng key? ──► Trả kết quả cũ đã lưu
   └───────────┬───────────┘
               │ Chưa có
               ▼
   ┌───────────────────────┐
   │  Role & Auth Check    │ ─── Không phải Admin? ──► HTTP 403
   └───────────┬───────────┘
               │ Hợp lệ
               ▼
   ┌───────────────────────┐
   │ Validate Rules        │ ─── BR-PB-02 (Trùng name cùng cha) / BR-PB-03 (>5 cấp)
   └───────────┬───────────┘
               │ Hợp lệ
               ▼
   ┌──────────────────────────────────────────────────────────┐
   │ DynamoDB TransactWriteItems:                             │
   │  1. Put Department (Condition: attribute_not_exists(PK)) │
   │  2. Put DEPT_CODE#<code> (Bảo vệ tính duy nhất code)      │
   │  3. Put AUDIT#<deptId> (Ghi nhật ký kiểm toán)           │
   │  4. Put IDEMPOTENCY#<key> (Lưu kết quả phản hồi)         │
   └───────────────────────────┬──────────────────────────────┘
                               │ Thành công
                               ▼
                        HTTP 201 Created
```

---

# 6. Thiết Kế Dữ Liệu DynamoDB (Data Model Changes)

Trong bảng `mini_hrm_table`, một thao tác tạo phòng ban sẽ ghi đồng thời 4 mục trong 1 transaction:

1. **Bản ghi Phòng ban chính**:
   - `PK`: `DEPT#<id>`
   - `SK`: `METADATA`
   - `GSI1PK`: `DEPT_PARENT#<parentId | ROOT>`
   - `GSI1SK`: `PATH#<path>`
   - Thuộc tính: `id`, `code`, `name`, `parentId`, `path`, `managerId = ""`, `status = "ACTIVE"`, `version = 1`, `createdAt`.

2. **Bản ghi khóa duy nhất cho Mã (Unique Code Item)**:
   - `PK`: `DEPT_CODE#<code>`
   - `SK`: `METADATA`
   - `ConditionExpression`: `attribute_not_exists(PK)` -> Nếu mã đã có, giao dịch thất bại ngay lập tức (`TransactionCanceledException`).

3. **Bản ghi Nhật ký kiểm toán**:
   - `PK`: `AUDIT#<id>`
   - `SK`: `LOG#<occurredAt>#<auditId>`
   - Thuộc tính: `actorId`, `actorName`, `action = "department.created"`, `targetType = "Department"`, `targetId = <id>`, `before = null`, `after = { code, name, parentId, path }`, `occurredAt`.

4. **Bản ghi Idempotency**:
   - `PK`: `IDEMPOTENCY#<Idempotency-Key>`
   - `SK`: `METADATA`
   - `TTL`: Thời gian hết hạn sau 24 giờ.

---

# 7. Thiết Kế API (API Design)

- **Endpoint**: `POST /api/v1/departments`
- **Headers**:
  - `Content-Type: application/json`
  - `Idempotency-Key: <UUIDv4>` (Bắt buộc)
  - Cookie: `session_token=<JWT>`

- **Request Body**:
  ```json
  {
    "code": "TECH_BE",
    "name": "Phòng Backend",
    "parentId": "dept_tech_001"
  }
  ```

- **Response Thành Công (`HTTP 201 Created`)**:
  ```json
  {
    "success": true,
    "data": {
      "id": "dept_be_101",
      "code": "TECH_BE",
      "name": "Phòng Backend",
      "parentId": "dept_tech_001",
      "path": "/dept_root_01/dept_tech_001/dept_be_101/",
      "managerId": "",
      "status": "ACTIVE",
      "version": 1
    }
  }
  ```

- **Response Lỗi Trùng Mã (`HTTP 409 Conflict` / `400 Bad Request`)**:
  ```json
  {
    "success": false,
    "error": {
      "code": "DUPLICATE_DEPARTMENT_CODE",
      "message": "Mã phòng ban TECH_BE đã tồn tại trong hệ thống. Vui lòng chọn mã khác!"
    }
  }
  ```

---

# 8. Xử Lý Lỗi (Failure Handling)
- Nếu DynamoDB báo lỗi `TransactionCanceledException` do điều kiện mã phòng trùng: Bắt lỗi và trả về thông điệp rõ ràng cho client.
- Nếu phòng cha `parentId` không tồn tại hoặc có `status = "ARCHIVED"`: Trả về `400 Bad Request` kèm thông điệp tương ứng.
- Nếu độ sâu vượt quá 5: Trả về `400 Bad Request` (*"Cây phòng ban đạt giới hạn tối đa 5 cấp"*).

---

# 9. Cân Nhắc Bảo Mật & Phân Quyền
- Kiểm tra vai trò trong JWT: Chỉ chấp nhận `role == "admin"`. Trưởng phòng hoặc nhân viên gọi endpoint này sẽ bị từ chối `HTTP 403 Forbidden`.

---

# 10. Chiến Lược Kiểm Thử (Testing Strategy)
- Unit Test Go cho hàm tính toán `path`: Các trường hợp phòng gốc, phòng con cấp 2, 3, 4, 5.
- Integration Test với DynamoDB Local:
  - Test tạo thành công.
  - Test trùng code -> Rollback toàn bộ, không có dòng Audit nào bị lọt.
  - Test gửi lại cùng `Idempotency-Key` -> Không tạo bản ghi thứ hai.
