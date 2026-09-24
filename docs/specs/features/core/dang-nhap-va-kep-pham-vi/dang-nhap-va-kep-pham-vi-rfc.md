# Đề Xuất Kỹ Thuật (RFC) — Đăng Nhập & Kẹp Phạm Vi Dữ Liệu

## 1. Kiến Trúc Xác Thực & Quản Lý Phiên (Cognito & HttpOnly Cookie)

```
[Trình duyệt] ── (POST /api/bff/auth/login) ──► [Next.js BFF]
                                                      │ Gọi Go API / Cognito
                                                      ▼
[Trình duyệt] ◄── Set-Cookie: session_token=JWT; ── [Next.js BFF]
                  HttpOnly; Secure; SameSite=Strict
```

- Token JWT **không bao giờ trả về JSON body** cho JavaScript trình duyệt.
- Tầng Next.js BFF đóng vai trò Proxy an toàn chuyển tiếp token sang Go API qua Header `Authorization: Bearer <token>`.

---

## 2. Kiến Trúc Kẹp Phạm Vi Truy Vấn (Data Scoping Engine)

### 2.1. Giải Mã Scope Động Trong Go Middleware
1. Middleware đọc JWT, lấy `userId` và `role`.
2. Nếu `role == "admin"`: Scope = `SCOPE_ALL`.
3. Nếu `role == "manager"`:
   - Đọc thông tin phòng ban hiện tại của Trưởng phòng từ DynamoDB (bộ nhớ đệm In-memory cache với TTL = 60 giây).
   - Lấy `deptPath` của phòng đó. Scope = `SCOPE_SUBTREE(deptPath)`.
4. Nếu `role == "employee"`:
   - Scope = `SCOPE_SELF(userId)`.

### 2.2. Kẹp Phạm Vi Bằng Truy Vấn DynamoDB (Ràng Buộc 1)
- Khi Trưởng phòng xem danh sách nhân viên trong cây con:
  ```go
  // Sử dụng GSI1 với truy vấn tiền tố path
  input := &dynamodb.QueryInput{
      TableName: aws.String("mini_hrm_table"),
      IndexName: aws.String("GSI1"),
      KeyConditionExpression: aws.String("GSI1PK = :deptPK"),
      ExpressionAttributeValues: map[string]types.AttributeValue{
          ":deptPK": &types.AttributeValueMemberS{Value: "DEPT#" + currentDeptId},
      },
  }
  ```
- Tuyệt đối không thực hiện `Scan` toàn bộ bảng rồi lọc trong RAM máy chủ.

### 2.3. Quy Tắc Bảo Mật 404 (Ràng Buộc 2)
```go
emp, err := repo.GetEmployeeByID(ctx, requestedEmpId)
if err != nil || !scope.CanAccess(emp) {
    // Trả về 404 hệt như bản ghi không tồn tại
    return RenderError(w, http.StatusNotFound, "Không tìm thấy hồ sơ nhân viên")
}
```
