# Yêu Cầu Sản Phẩm (PRD) — Đăng Nhập & Kẹp Phạm Vi Dữ Liệu (Auth & Data Scoping)

## Mục Tiêu
PRD định nghĩa **CÁI GÌ** và **TẠI SAO** cần xây dựng cơ chế Định danh, Xác thực qua Cognito và Kẹp phạm vi truy vấn dữ liệu động trong Mini HRM.

---

# 1. Vấn Đề (Problem Statement)
- Hệ thống quản lý thông tin nhân sự và tổ chức chứa nhiều dữ liệu nhạy cảm.
- Nếu token xác thực được lưu ở `localStorage` hoặc biến JavaScript, nguy cơ bị rò rỉ qua các cuộc tấn công XSS là rất lớn.
- Nếu phạm vi quyền hạn (như danh sách phòng ban được quản lý) bị "đóng băng" cứng vào trong JWT Token lúc đăng nhập: Khi Quản trị viên đổi phòng của một Trưởng phòng, người đó vẫn giữ nguyên quyền cũ cho tới khi token hết hạn hoặc phải ép đăng nhập lại.
- Nếu lọc dữ liệu sau khi truy vấn về RAM: Dữ liệu ngoài quyền đã rời DB, dễ rò rỉ qua log và làm chậm hệ thống.
- Nếu trả về `403 Forbidden` khi xem bản ghi ngoài quyền: Kẻ xấu có thể dò quét (brute-force) để lập danh sách toàn bộ ID và hồ sơ thật trong công ty.

---

# 2. Mục Tiêu (Goals)
- Tích hợp xác thực qua AWS Cognito, bảo mật phiên tuyệt đối bằng Cookie `HttpOnly`.
- Triển khai phân cấp 3 vai trò: Quản trị viên, Trưởng phòng, Nhân viên.
- Thực hiện **Kẹp phạm vi động từ CSDL (Data Scoping)**: Đổi phòng là đổi phạm vi trong vòng tối đa 60 giây (`BR-PV-04`).
- Kẹp phạm vi là điều kiện truy vấn trực tiếp tại CSDL (Ràng buộc 1).
- Trả về `404 Not Found` thay vì `403 Forbidden` cho các bản ghi ngoài phạm vi (Ràng buộc 2).

---

# 3. Yêu Cầu Chức Năng & Quy Tắc
- **FR-SEC-01**: Đăng nhập qua Cognito, máy chủ Next.js thiết lập cookie `HttpOnly`.
- **FR-SEC-02**: Quản trị viên truy vấn toàn bộ dữ liệu công ty (`BR-PV-01`).
- **FR-SEC-03**: Trưởng phòng chỉ truy vấn được phòng mình và các phòng con (`BR-PV-02`).
- **FR-SEC-04**: Nhân viên chỉ truy vấn được hồ sơ chính mình (`BR-PV-03`).
- **FR-SEC-05**: Hồ sơ ngoài phạm vi phải trả về HTTP `404 Not Found`.
