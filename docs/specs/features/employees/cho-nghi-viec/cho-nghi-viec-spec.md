# Đặc Tả Hành Vi Tính Năng (Feature Spec) — Cho Nhân Viên Nghỉ Việc

## 1. Tiêu Chí Nghiệm Thu (Acceptance Criteria)
- [ ] Nhân viên đang là Trưởng phòng -> Bị từ chối, thông báo chỉ rõ phòng ban nào và yêu cầu gỡ chức vụ trước.
- [ ] Nhân viên bình thường -> Cập nhật thành công sang trạng thái `RESIGNED`.
- [ ] Sau khi nghỉ việc: Nhân viên không còn xuất hiện trong danh sách nhân sự mặc định (chỉ xuất hiện khi bật bộ lọc *"Đã nghỉ việc"*).
- [ ] Email của nhân viên đã nghỉ việc: Thử tạo nhân viên mới với email này -> Bị từ chối (`BR-NV-01`).
- [ ] Bản ghi vẫn còn nguyên vẹn trong DynamoDB, không bị mất mát bất kỳ trường dữ liệu nào.
