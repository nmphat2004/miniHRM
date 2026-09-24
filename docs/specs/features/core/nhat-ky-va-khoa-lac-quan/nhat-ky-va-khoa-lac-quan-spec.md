# Đặc Tả Hành Vi Tính Năng (Feature Spec) — Nhật Ký & Khóa Lạc Quan

## 1. Tiêu Chí Nghiệm Thu (Acceptance Criteria)
- [ ] Hai Quản trị viên cùng sửa một phòng ban: Người bấm sau nhận mã lỗi **`409 Conflict`** (EC-01).
- [ ] Gửi lại cùng một `Idempotency-Key` trong vòng 24 giờ: Nhận lại kết quả cũ, không tạo thêm bản ghi thứ hai trong CSDL (EC-04).
- [ ] Thao tác nghiệp vụ bị lỗi/hủy: Tuyệt đối không có bất kỳ dòng nhật ký rác nào được ghi lại (Ràng buộc 3).
- [ ] Thao tác thành công: Có đúng một bản ghi `AuditLog` ghi rõ ai làm, lúc nào, trường nào thay đổi.
- [ ] Màn hình danh sách/chi tiết hiển thị đầy đủ 6 trạng thái UI theo chuẩn UX.
