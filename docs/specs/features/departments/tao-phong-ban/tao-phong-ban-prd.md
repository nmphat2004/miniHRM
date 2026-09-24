# Yêu Cầu Sản Phẩm (PRD) — Tính Năng Tạo Phòng Ban Mới

## Mục Tiêu
PRD định nghĩa **CÁI GÌ** và **TẠI SAO** cần xây dựng tính năng Tạo phòng ban mới trong Mini HRM. Tính năng này cho phép Quản trị viên thiết lập và mở rộng cơ cấu tổ chức hình cây của doanh nghiệp.

---

# 1. Vấn Đề (Problem Statement)
- Hiện tại khi doanh nghiệp mở rộng quy mô (từ 200 đến 500 nhân sự), việc thành lập thêm các khối, ban, phòng chuyên môn hoặc các tổ dự án mới diễn ra thường xuyên.
- Nếu không có cơ chế phân cấp cha–con rõ ràng, cơ cấu tổ chức sẽ bị phẳng hóa, gây khó khăn cho việc quản lý nhân sự và phân quyền dữ liệu.
- Nguy cơ trùng lặp mã phòng ban hoặc tạo phòng ban vượt quá giới hạn độ sâu cho phép sẽ làm vỡ giao diện sơ đồ cây và phá hỏng cơ chế kẹp phạm vi (Data Scoping).

---

# 2. Mục Tiêu (Goals)
- Cung cấp giao diện trực quan cho Quản trị viên khởi tạo phòng ban mới (phòng gốc hoặc phòng con).
- Đảm bảo tính duy nhất của mã phòng ban (`code`) trên toàn công ty.
- Đảm bảo tính duy nhất của tên phòng ban (`name`) trong cùng một cấp cha.
- Tự động kiểm soát chiều sâu cây tổ chức không vượt quá **5 cấp**.
- Tự động sinh chuỗi định danh phân cấp (`path`) hỗ trợ truy vấn cây con.

---

# 3. Không Nằm Trong Phạm Vi (Non-Goals)
- Không cho phép Trưởng phòng hoặc Nhân viên tự tạo phòng ban (chỉ Quản trị viên).
- Không tự động bổ nhiệm Trưởng phòng ngay trong luồng tạo phòng ban (thực hiện qua luồng bổ nhiệm riêng hoặc cập nhật sau).
- Không hỗ trợ nhập hàng loạt qua Excel trong giai đoạn này.

---

# 4. Người Dùng Mục Tiêu (Target Users)
- **Quản trị viên (Admin)**: Người duy nhất được cấp quyền tạo lập cơ cấu tổ chức.

---

# 5. Hành Trình Người Dùng (User Journey)
1. Quản trị viên truy cập màn hình *"Sơ đồ phòng ban"*.
2. Bấm nút *"Thêm phòng ban mới"* (tại nút gốc hoặc tại một nút phòng ban cụ thể trên cây).
3. Nhập các trường thông tin: Mã phòng ban, Tên phòng ban, chọn Phòng ban cha (nếu có).
4. Bấm *"Xác nhận tạo"*.
5. Hệ thống kiểm tra hợp lệ và thông báo thành công. Phòng ban mới lập tức xuất hiện đúng vị trí trên cây tổ chức.

---

# 6. Yêu Cầu Chức Năng (Functional Requirements)
- **FR-PB-01**: Cho phép tạo phòng ban gốc (khi không chọn phòng ban cha).
- **FR-PB-02**: Cho phép tạo phòng ban con trực thuộc một phòng cha đang hoạt động (`ACTIVE`).
- **FR-PB-03**: Từ chối tạo phòng ban nếu mã (`code`) đã tồn tại trên toàn công ty (`BR-PB-01`).
- **FR-PB-04**: Từ chối tạo phòng ban nếu tên (`name`) đã tồn tại dưới cùng một phòng cha (`BR-PB-02`).
- **FR-PB-05**: Cho phép trùng tên (`name`) nếu nằm ở hai phòng cha khác nhau.
- **FR-PB-06**: Chặn tạo phòng ban con nếu phòng cha đã ở cấp thứ 5 (`BR-PB-03`).
- **FR-PB-07**: Tự động sinh đường dẫn `path` chuẩn xác: `parent.path + id + "/`.
- **FR-PB-08**: Ghi nhận một bản ghi nhật ký kiểm toán (`action = "department.created"`) đồng thời trong cùng transaction.

---

# 7. Quy Tắc Nghiệp Vụ (Business Rules)
- **BR-PB-01**: Mã phòng ban (`code`) duy nhất toàn công ty và bất biến sau khi tạo.
- **BR-PB-02**: Tên phòng ban (`name`) từ 2–100 ký tự, duy nhất trong cùng một phòng cha.
- **BR-PB-03**: Cây sâu tối đa 5 cấp.
- **BR-PB-07**: Phòng ban `ARCHIVED` không được tiếp nhận thêm phòng con mới.
- **Ràng buộc 3**: Ghi nhật ký cùng transaction với ghi nghiệp vụ.
- **Ràng buộc 6**: Nhận và xử lý `Idempotency-Key` chống tạo bản ghi trùng lặp khi bấm đúp.

---

# 8. Chỉ Số Thành Công (Success Metrics)
- 100% các phòng ban mới tạo có cấu trúc `path` chính xác.
- Tốc độ tạo và cập nhật cây trên giao diện dưới 500ms.
- Tỷ lệ trùng mã hoặc vượt quá 5 cấp độ sâu bằng 0.

---

# 9. Ràng Buộc (Constraints)
- Backend: Go HTTP Server / AWS Lambda.
- Database: Amazon DynamoDB (Single-Table Design).
- Idempotency: Header `Idempotency-Key` bắt buộc cho yêu cầu POST.

---

# 10. Rủi Ro (Risks)
- Người dùng bấm tạo nhiều lần khi mạng chậm -> Cần bảo vệ bằng `Idempotency-Key`.
- Tạo phòng con dưới một phòng cha đang bị xóa mềm/lưu trữ cùng lúc -> Cần kiểm tra trạng thái phòng cha tại thời điểm ghi.

---

# 11. Phụ Thuộc (Dependencies)
- Phân hệ Định danh & Xác thực (Xác định vai trò Quản trị viên).
- Bảng DynamoDB và partition lưu trữ khóa duy nhất `DEPT_CODE#<code>`.

---

# 12. Câu Hỏi Còn Mở (Open Questions)
- Không có. Quy tắc nghiệp vụ đã được chốt rõ tại đề bài Mini HRM (UC-01, BR-PB-01..03).
