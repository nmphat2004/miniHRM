# Đặc Tả Hành Vi Tính Năng (Feature Spec) — Đăng Nhập & Kẹp Phạm Vi

## 1. Tiêu Chí Nghiệm Thu (Acceptance Criteria)
- [ ] Cookie phiên có đủ cờ `HttpOnly`, `Secure`, `SameSite=Strict`. JavaScript gọi `document.cookie` không đọc được token.
- [ ] Quản trị viên đăng nhập: Xem được toàn bộ phòng ban và nhân sự công ty.
- [ ] Trưởng phòng đăng nhập: Chỉ thấy cây con của mình; gọi API lấy nhân viên phòng nhánh khác -> Nhận danh sách rỗng.
- [ ] Xem hồ sơ nhân viên phòng khác bằng ID: Nhận mã lỗi **`404 Not Found`**, tuyệt đối không phải `403`.
- [ ] Đổi phòng ban của Trưởng phòng: Trong vòng 60 giây, gọi lại API sẽ tự động nhận dữ liệu theo phạm vi phòng mới mà không cần đăng nhập lại.
- [ ] Nhân viên thường gọi thẳng API danh sách nhân sự: Nhận `403 Forbidden` hoặc chỉ nhận duy nhất hồ sơ của chính mình.
