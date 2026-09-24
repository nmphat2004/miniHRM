# Đặc Tả Hành Vi Tính Năng (Feature Spec) — Chuyển Phòng Ban Cho Nhân Viên

## 1. User Stories
> Là một **Quản trị viên**,  
> Tôi muốn **chuyển kỹ sư Nguyễn Văn A từ Phòng Frontend sang Phòng Mobile**,  
> Để **hỗ trợ kịp tiến độ dự án mới của công ty**.

---

# 2. Tiêu Chí Nghiệm Thu (Acceptance Criteria)
- [ ] Chuyển vào một phòng ban đang ở trạng thái `ARCHIVED` -> Bị từ chối (`BR-PB-07`).
- [ ] Nhân viên đang là Trưởng phòng cũ: Hiển thị cảnh báo rõ ràng; nếu Quản trị viên xác nhận thì tự động gỡ chức vụ Trưởng phòng cũ.
- [ ] Bản ghi nhật ký `AuditLog` có đủ thông tin ID phòng cũ và ID phòng mới (`BR-NV-05`).
- [ ] Trưởng phòng cũ và Trưởng phòng mới mở lại danh sách nhân sự của mình: Danh sách tự động cập nhật ngay lập tức.
