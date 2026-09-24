# Đề Xuất Kỹ Thuật (RFC) — Cho Nhân Viên Nghỉ Việc

## Mục Tiêu
Thiết kế giải pháp kỹ thuật, API và cơ chế bảo toàn dữ liệu cho quy trình thôi việc (`UC-04`).

---

# 1. Kiến Trúc Kỹ Thuật & DynamoDB Design

### 1.1. Kiểm tra ràng buộc Trưởng phòng (`BR-NV-04`)
Trước khi cập nhật trạng thái, Service thực hiện kiểm tra:
```go
// Quét xem employeeId có đang được gán làm managerId của bất kỳ phòng ban nào không
dept, isManager := deptRepo.FindDepartmentByManagerID(ctx, empId)
if isManager {
    return fmt.Errorf("ERR_EMPLOYEE_IS_ACTIVE_MANAGER: nhân viên đang là Trưởng phòng của %s (id: %s). Vui lòng gỡ chức vụ trước!", dept.Name, dept.ID)
}
```

### 1.2. Thao tác ghi TransactWriteItems
Giao dịch cập nhật bao gồm:
1. **Cập nhật trạng thái Employee**:
   - `PK = "EMP#<id>"`, `SK = "METADATA"`
   - `UpdateExpression`: `SET #st = :resigned, resignedAt = :date, version = version + 1`
   - `ExpressionAttributeNames`: `{"#st": "status"}`
   - `ConditionExpression`: `#st = :active AND version = :expectedVersion`
2. **Ghi nhật ký kiểm toán**:
   - `PK = "AUDIT#<id>"`, `SK = "LOG#<occurredAt>#<auditId>"`
   - `Item`: `action = "employee.resigned"`, `before = { status: "ACTIVE" }`, `after = { status: "RESIGNED", resignedAt: :date }`.

> **Lưu ý về Email**: Bản ghi `EMP_EMAIL#<email>` **tuyệt đối không bị xóa**. Do đó bất kỳ lời gọi tạo nhân viên mới nào với email này trong tương lai đều sẽ bị DynamoDB từ chối vì trùng khóa PK.

---

# 2. Thiết Kế API
- **Endpoint**: `POST /api/v1/employees/{id}/resign`
- **Headers**:
  - `Idempotency-Key: <UUIDv4>`
- **Request Body**:
  ```json
  {
    "resignedAt": "2026-09-30",
    "reason": "Chuyển định cư"
  }
  ```
- **Response (`HTTP 200 OK`)**:
  ```json
  {
    "success": true,
    "data": {
      "id": "emp_456",
      "status": "RESIGNED",
      "resignedAt": "2026-09-30"
    }
  }
  ```
