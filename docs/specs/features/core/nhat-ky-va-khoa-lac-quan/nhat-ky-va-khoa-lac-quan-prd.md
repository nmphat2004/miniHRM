# Yêu Cầu Sản Phẩm (PRD) — Nhật Ký Kiểm Toán, Khóa Lạc Quan & Idempotency

## Mục Tiêu
PRD định nghĩa **CÁI GÌ** và **TẠI SAO** cần xây dựng hệ thống kiểm toán, tính toàn vẹn giao dịch và cơ chế phòng vệ chống lỗi thao tác trong Mini HRM.

---

# 1. Vấn Đề (Problem Statement)
- **Mất dấu vết khi có sự cố**: Nếu nhật ký kiểm toán ghi sau giao dịch, có những lúc dữ liệu đã thay đổi nhưng nhật ký bị lỗi mạng không ghi được. Khi xảy ra tranh chấp hoặc lỗi hệ thống, không ai biết ai đã thay đổi dữ liệu đó.
- **Xung đột ghi đè ngầm (Lost Updates)**: Khi hai Quản trị viên cùng mở màn hình chỉnh sửa một phòng ban, người lưu sau sẽ âm thầm đè bẹp thay đổi của người lưu trước mà không có bất kỳ cảnh báo nào.
- **Tạo bản ghi rác khi mạng lag**: Người dùng sốt ruột bấm 2–3 lần nút "Lưu", dẫn tới tạo nhiều bản ghi trùng lặp.
- **Giao diện thiếu trạng thái**: Giao diện chỉ xử lý trạng thái Đang tải và Thành công; bỏ quên trạng thái Mất kết nối và Không có quyền khiến người dùng hoang mang khi mạng chập chờn.

---

# 2. Mục Tiêu (Goals)
- Đảm bảo nhật ký kiểm toán luôn được ghi nguyên tử cùng nghiệp vụ (Ràng buộc 3).
- Ngăn chặn triệt để hiện tượng ghi đè ngầm qua Khóa lạc quan (Ràng buộc 5).
- Đảm bảo tính lũy kế (Idempotency) cho mọi yêu cầu POST (Ràng buộc 6).
- Chuẩn hóa trải nghiệm người dùng với đầy đủ 6 trạng thái giao diện (Ràng buộc 8).

---

# 3. Yêu Cầu Chức Năng & Quy Tắc
- **FR-AUD-01**: Sổ nhật ký chỉ ghi thêm (Append-only), không cho phép sửa/xóa.
- **FR-AUD-02**: Chụp ảnh `before` và `after` chỉ lưu lại các trường có thay đổi thực tế.
- **FR-AUD-03**: Mọi thao tác cập nhật dữ liệu phải kiểm tra số phiên bản `version`. Nếu không khớp, từ chối với HTTP `409 Conflict`.
- **FR-AUD-04**: Tiếp nhận Header `Idempotency-Key`. Nếu nhận lại cùng một key, trả về kết quả cũ đã lưu.
- **FR-AUD-05**: Mọi màn hình phải có đầy đủ giao diện cho 6 trạng thái: Đang tải, Rỗng, Lỗi, Không có quyền, Đang lưu, Mất kết nối.
