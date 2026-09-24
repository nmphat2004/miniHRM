# Yêu Cầu Sản Phẩm (PRD) — Chuyển Phòng Ban Cho Nhân Viên (Transfer Employee)

## Mục Tiêu
PRD định nghĩa **CÁI GÌ** và **TẠI SAO** cần xây dựng tính năng Điều chuyển phòng ban cho nhân sự trong Mini HRM. Tính năng này cho phép Quản trị viên luân chuyển nhân sự giữa các đơn vị phòng ban khác nhau trong công ty.

---

# 1. Vấn Đề (Problem Statement)
- Nhân sự trong doanh nghiệp thường xuyên được điều chuyển công tác, luân chuyển phòng ban để đáp ứng nhu cầu dự án.
- **Rủi ro bảo mật trọng yếu (Leo thang quyền)**: Nếu cấp quyền chuyển phòng ban cho Trưởng phòng, họ có thể lén chuyển nhân viên của phòng khác vào cây con của mình rồi sửa hồ sơ. Do đó, luồng này bắt buộc phải khóa cứng chỉ cho Quản trị viên.
- Nếu chuyển nhân sự vào một phòng ban đã giải thể/lưu trữ (`ARCHIVED`), nhân sự đó sẽ rơi vào trạng thái "ma" không người quản lý.
- Nếu nhân sự đang là Trưởng phòng của phòng cũ, việc chuyển đi mà không cảnh báo sẽ để lại một phòng ban hoạt động với người đứng đầu không còn ở đó.

---

# 2. Mục Tiêu (Goals)
- Cho phép Quản trị viên chuyển một nhân viên từ phòng ban hiện tại sang một phòng ban mới đang hoạt động (`ACTIVE`).
- Cảnh báo và yêu cầu xác nhận bắt buộc nếu nhân sự đang giữ vai trò Trưởng phòng ở đơn vị cũ.
- Tự động phản ánh danh sách nhân sự mới cho cả Trưởng phòng cũ và Trưởng phòng mới trong vòng 60 giây.
- Ghi nhận nhật ký kiểm toán minh bạch ghi rõ từ phòng nào sang phòng nào (`BR-NV-05`).

---

# 3. Không Nằm Trong Phạm Vi (Non-Goals)
- **Tuyệt đối không cấp quyền cho Trưởng phòng** (ngăn chặn lỗ hổng leo thang đặc quyền).
- Không tự động thay đổi mức lương hay chức danh (sử dụng chức năng sửa hồ sơ riêng).

---

# 4. Người Dùng Mục Tiêu (Target Users)
- **Quản trị viên (Admin)**: Người duy nhất có quyền điều chuyển.

---

# 5. Hành Trình Người Dùng (User Journey)
1. Quản trị viên vào hồ sơ nhân viên hoặc danh sách nhân sự.
2. Chọn thao tác *"Chuyển phòng ban"*.
3. Chọn phòng ban mới từ danh sách các phòng đang `ACTIVE`.
4. Nếu nhân viên đang là Trưởng phòng: Hệ thống hiển thị hộp thoại cảnh báo: *"Nhân viên này đang là Trưởng phòng [Tên phòng]. Chuyển phòng sẽ tự động hủy vai trò Trưởng phòng cũ. Bạn có chắc chắn không?"*.
5. Quản trị viên xác nhận.
6. Hệ thống cập nhật dữ liệu và hiển thị thông báo thành công.

---

# 6. Yêu Cầu Chức Năng (Functional Requirements)
- **FR-NV-10**: Chỉ Quản trị viên được gọi chức năng này.
- **FR-NV-11**: Từ chối chuyển nếu phòng ban mới đang `ARCHIVED` (`BR-PB-07`).
- **FR-NV-12**: Kiểm tra và cảnh báo nếu nhân viên đang là Trưởng phòng ở đơn vị hiện tại.
- **FR-NV-13**: Cập nhật trường `departmentId` của nhân viên.
- **FR-NV-14**: Ghi một bản ghi nhật ký `AuditLog` nêu rõ phòng ban cũ và mới (`BR-NV-05`).

---

# 7. Quy Tắc Nghiệp Vụ (Business Rules)
- **BR-PB-07**: Phòng ban `ARCHIVED` không tiếp nhận nhân viên mới.
- **BR-NV-05**: Chuyển phòng ban phải ghi nhật ký nêu rõ từ phòng nào sang phòng nào.
- **BR-PV-04**: Phạm vi tính từ CSDL, Trưởng phòng cũ và mới thấy danh sách thay đổi ngay lập tức.
