# Yêu Cầu Sản Phẩm (PRD) — Cho Nhân Viên Nghỉ Việc (Employee Resignation)

## Mục Tiêu
PRD định nghĩa **CÁI GÌ** và **TẠI SAO** cần xây dựng quy trình Cho nhân viên nghỉ việc (xóa mềm) trong Mini HRM (`UC-04`).

---

# 1. Vấn Đề (Problem Statement)
- Nhân sự thôi việc là quy trình bình thường trong doanh nghiệp. Tuy nhiên nếu xóa cứng bản ghi (Hard Delete) khỏi cơ sở dữ liệu:
  - Toàn bộ lịch sử công tác, vết chấm công, nhật ký phê duyệt đơn từ của nhân sự đó trong quá khứ sẽ bị bốc hơi hoàn toàn.
  - Các liên kết khóa ngoại và nhật ký kiểm toán sẽ bị mồ côi (Dangling References).
- **Rủi ro rò rỉ và tranh chấp lịch sử**: Nếu địa chỉ email của người nghỉ việc được giải phóng cho nhân sự mới sử dụng lại, toàn bộ thư từ và vết lịch sử trong quá khứ sẽ bị quy nhầm cho người mới.
- **Rủi ro người đứng đầu ảo**: Nếu cho phép một nhân viên đang là Trưởng phòng nghỉ việc mà không có người thay thế, phòng ban đó sẽ trở thành đơn vị không có người quản trị thực tế.

---

# 2. Mục Tiêu (Goals)
- Cho phép Quản trị viên cập nhật trạng thái nghỉ việc (`status = "RESIGNED"`) cho nhân sự.
- Ngăn chặn triệt để thao tác nếu nhân sự đang giữ vai trò Trưởng phòng (`BR-NV-04`).
- Đảm bảo bản ghi được bảo toàn nguyên vẹn trong CSDL (Xóa mềm - `BR-NV-03`).
- Khóa vĩnh viễn địa chỉ email, không cấp lại cho bất kỳ ai (`BR-NV-01`).
- Ẩn nhân viên đã nghỉ việc khỏi danh sách nhân sự mặc định (chỉ hiển thị khi bật bộ lọc riêng).

---

# 3. Yêu Cầu Chức Năng (Functional Requirements)
- **FR-NV-20**: Duy nhất Quản trị viên được thực hiện thao tác.
- **FR-NV-21**: Kiểm tra nếu nhân viên đang là Trưởng phòng của bất kỳ phòng ban nào -> **Chặn đứng và nêu rõ phòng ban liên quan**.
- **FR-NV-22**: Cập nhật trạng thái thành `RESIGNED`, ghi nhận ngày nghỉ việc.
- **FR-NV-23**: Giữ nguyên bản ghi bảo vệ `EMP_EMAIL#<email>` trong DynamoDB để ngăn tái sử dụng.
- **FR-NV-24**: Ghi một bản ghi nhật ký `AuditLog` (`action = "employee.resigned"`).

---

# 4. Quy Tắc Nghiệp Vụ (Business Rules)
- **BR-NV-01**: `email` duy nhất toàn công ty, kể cả hồ sơ đã nghỉ việc.
- **BR-NV-03**: Nghỉ việc là xóa mềm — đổi `status`, giữ nguyên bản ghi và lịch sử.
- **BR-NV-04**: Nhân viên đang là Trưởng phòng không được phép nghỉ việc cho tới khi chuyển giao chức vụ.
