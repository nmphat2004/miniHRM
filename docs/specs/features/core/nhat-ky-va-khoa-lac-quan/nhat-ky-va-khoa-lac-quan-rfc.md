# Đề Xuất Kỹ Thuật (RFC) — Nhật Ký Kiểm Toán & Khóa Lạc Quan

## 1. Cơ Chế Khóa Lạc Quan (Optimistic Concurrency Control - OCC)

### Lệnh cập nhật có điều kiện trên DynamoDB:
```text
UpdateItem:
  Key: { PK: "DEPT#123", SK: "METADATA" }
  UpdateExpression: "SET #name = :newName, version = version + :inc"
  ConditionExpression: "attribute_exists(PK) AND version = :expectedVersion"
  ExpressionAttributeValues: {
      ":expectedVersion": 2,
      ":newName": "Tên mới",
      ":inc": 1
  }
```
- Nếu phiên bản trong DB đã là `3`, DynamoDB ném `ConditionalCheckFailedException`.
- Tầng Go API bắt ngoại lệ này và trả về mã lỗi HTTP **`409 Conflict`** kèm thông điệp rõ ràng cho client.

---

## 2. Cơ Chế Idempotency Cho Thao Tác POST

### Bảng lưu vết Idempotency trong DynamoDB:
- `PK`: `IDEMPOTENCY#<Idempotency-Key>`
- `SK`: `METADATA`
- `status`: `"PROCESSING"` | `"COMPLETED"`
- `responseCode`: `201`
- `responseBody`: JSON kết quả đã xử lý
- `ttl`: Unix timestamp (hết hạn sau 24 giờ)

### Middleware xử lý trong Go:
```go
func IdempotencyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            next.ServeHTTP(w, r)
            return
        }
        key := r.Header.Get("Idempotency-Key")
        if key == "" {
            next.ServeHTTP(w, r)
            return
        }
        // Kiểm tra xem key đã tồn tại chưa
        record, exists := repo.GetIdempotency(r.Context(), key)
        if exists && record.Status == "COMPLETED" {
            // Trả ngay kết quả cũ
            w.WriteHeader(record.ResponseCode)
            w.Write(record.ResponseBody)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## 3. Ghi Nhật Ký Kiểm Toán Nguyên Tử (Atomic Audit Logging)
Mọi lệnh cập nhật hoặc tạo mới đều được thực thi qua hàm:
```go
func TransactWithAudit(ctx context.Context, businessWriteItem types.TransactWriteItem, auditLog *AuditLog) error {
    auditWriteItem := types.TransactWriteItem{
        Put: &types.Put{
            TableName: aws.String("mini_hrm_table"),
            Item: marshalAuditLog(auditLog),
        },
    }
    _, err := dynamoClient.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
        TransactItems: []types.TransactWriteItem{businessWriteItem, auditWriteItem},
    })
    return err
}
```
