# Đề Xuất Kỹ Thuật (RFC) — Chuyển Phòng Ban Cho Nhân Viên

## Mục Tiêu
Thiết kế giải pháp kỹ thuật, API và giao dịch DynamoDB cho thao tác điều chuyển phòng ban (`UC-03`).

---

# 1. Kiến Trúc Giao Dịch DynamoDB (Atomic Transaction)

Khi điều chuyển nhân viên `emp_123` từ phòng `dept_old` sang phòng `dept_new`:
Giao dịch `TransactWriteItems` bao gồm 3 hành động:

1. **Cập nhật bản ghi Nhân viên**:
   - `Key`: `PK = "EMP#emp_123"`, `SK = "METADATA"`
   - `UpdateExpression`: `SET departmentId = :newDept, GSI1PK = :newDeptPK, version = version + 1`
   - `ConditionExpression`: `version = :currentVersion` (Khóa lạc quan OCC).
2. **Nếu nhân viên là Trưởng phòng cũ (gỡ bỏ managerId)**:
   - `Key`: `PK = "DEPT#dept_old"`, `SK = "METADATA"`
   - `UpdateExpression`: `SET managerId = :empty, version = version + 1`
   - `ConditionExpression`: `managerId = :empId`
3. **Ghi nhật ký kiểm toán**:
   - `Key`: `PK = "AUDIT#emp_123"`, `SK = "LOG#<occurredAt>#<id>"`
   - `Item`: `action = "employee.transferred"`, `before = { departmentId: "dept_old" }`, `after = { departmentId: "dept_new" }`.

---

# 2. Thiết Kế API
- **Endpoint**: `POST /api/v1/employees/{id}/transfer`
- **Headers**:
  - `Idempotency-Key: <UUIDv4>`
- **Request Body**:
  ```json
  {
    "newDepartmentId": "dept_tech_be",
    "confirmManagerRemoval": true
  }
  ```
- **Response (`HTTP 200 OK`)**:
  ```json
  {
    "success": true,
    "data": {
      "id": "emp_123",
      "oldDepartmentId": "dept_mkt",
      "newDepartmentId": "dept_tech_be",
      "transferredAt": "2026-09-23T07:00:00Z"
    }
  }
  ```
