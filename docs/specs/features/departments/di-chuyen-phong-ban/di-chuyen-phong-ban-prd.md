# Yêu Cầu Sản Phẩm (PRD) — Di Chuyển Phòng Ban (Move Department)

## Mục Tiêu
PRD định nghĩa **CÁI GÌ** và **TẠI SAO** cần xây dựng tính năng Di chuyển phòng ban trong Mini HRM. Đây là **luồng nghiệp vụ phức tạp và rủi ro nhất toàn hệ thống**, cho phép tái cấu trúc sơ đồ tổ chức doanh nghiệp khi có sự sáp nhập, chia tách hoặc điều chuyển đơn vị.

---

# 1. Vấn Đề (Problem Statement)
- Trong quá trình phát triển, doanh nghiệp thường xuyên tái cơ cấu: Chuyển một Tổ kỹ thuật từ Phòng Dự án sang Khối Công nghệ, hoặc sáp nhập các đơn vị trực thuộc.
- **Rủi ro chí mạng**:
  1. Nếu cho phép chuyển một phòng ban vào chính nó hoặc vào con cháu của nó, đồ thị cây sẽ bị biến thành chu trình lặp vô tận, làm sập toàn bộ ứng dụng.
  2. Việc di chuyển một phòng ban sẽ làm thay đổi đường dẫn phân cấp của toàn bộ phòng con bên dưới. Nếu cập nhật thiếu sót, dữ liệu cây sẽ bị đứt gãy.
  3. Di chuyển một phòng ban sẽ lập tức làm thay đổi **phạm vi quản lý dữ liệu của mọi Trưởng phòng cấp dưới**.

---

# 2. Mục Tiêu (Goals)
- Cho phép Quản trị viên di chuyển một phòng ban (kèm toàn bộ cây con) sang một phòng ban cha mới đang hoạt động.
- Ngăn chặn 100% nguy cơ tạo chu trình đệ quy (`BR-PB-04`).
- Đảm bảo độ sâu của nhánh con sâu nhất sau khi chuyển không vượt quá 5 cấp (`BR-PB-03`).
- Cập nhật chính xác và đồng bộ trường `path` cho toàn bộ cây con (`BR-PB-05`).
- Tự động cập nhật phạm vi truy cập dữ liệu của các Trưởng phòng bị ảnh hưởng.

---

# 3. Không Nằm Trong Phạm Vi (Non-Goals)
- Không cho phép Trưởng phòng tự ý di chuyển phòng ban (chỉ Quản trị viên).
- Không tự động tách rời nhân viên ra khỏi phòng ban khi di chuyển (nhân viên vẫn ở nguyên trong phòng ban của họ).

---

# 4. Người Dùng Mục Tiêu (Target Users)
- **Quản trị viên (Admin)**: Người duy nhất có thẩm quyền thay đổi cấu trúc cây công ty.

---

# 5. Hành Trình Người Dùng (User Journey)
1. Quản trị viên chọn phòng ban cần chuyển trên cây tổ chức.
2. Chọn hành động *"Di chuyển phòng ban"*.
3. Chọn phòng ban cha mới từ danh sách thả xuống (hoặc kéo thả trên sơ đồ).
4. Hệ thống hiển thị bản xem trước (Preview): Cảnh báo số lượng phòng con và nhân sự sẽ bị chuyển dời, xác nhận độ sâu mới hợp lệ.
5. Quản trị viên bấm *"Xác nhận di chuyển"*.
6. Hệ thống thực hiện cập nhật toàn bộ cây con và phản hồi kết quả thành công.

---

# 6. Yêu Cầu Chức Năng (Functional Requirements)
- **FR-PB-10**: Chỉ cho phép Quản trị viên thực hiện di chuyển.
- **FR-PB-11**: Từ chối di chuyển nếu phòng nguồn hoặc phòng cha mới đang ở trạng thái `ARCHIVED`.
- **FR-PB-12**: Phát hiện và chặn đứng việc di chuyển vào chính nó (`sourceId == targetParentId`).
- **FR-PB-13**: Phát hiện và chặn đứng việc di chuyển vào bất kỳ phòng ban con nào trong cây con của chính nó (`BR-PB-04`).
- **FR-PB-14**: Tính toán độ sâu nhánh lớn nhất và chặn thao tác nếu vượt quá 5 cấp (`BR-PB-03`).
- **FR-PB-15**: Cập nhật đồng bộ chuỗi dẫn xuất `path` cho phòng nguồn và toàn bộ phòng con ở mọi độ sâu (`BR-PB-05`).
- **FR-PB-16**: Ghi nhận một bản ghi `AuditLog` chi tiết thể hiện ID cha cũ, ID cha mới, path cũ và path mới.

---

# 7. Quy Tắc Nghiệp Vụ (Business Rules)
- **BR-PB-03**: Cây sâu tối đa 5 cấp.
- **BR-PB-04**: Không được di chuyển một phòng ban vào chính nó hoặc vào cây con của nó.
- **BR-PB-05**: Di chuyển phòng ban phải cập nhật path của toàn bộ cây con.
- **BR-PV-04**: Phạm vi quản lý của các Trưởng phòng bên dưới tự động đổi theo trong 60 giây.

---

# 8. Chỉ Số Thành Công (Success Metrics)
- Tỷ lệ xảy ra chu trình đệ quy = 0%.
- Toàn bộ cây con sau khi di chuyển có `path` khớp 100% với cấu trúc mới.
- Không có trường hợp cây bị rơi vào trạng thái nửa vời khi phát sinh lỗi giữa chừng.
