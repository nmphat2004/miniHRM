# Đặc Tả Hành Vi Tính Năng (Feature Spec) — Tạo Phòng Ban Mới

## Mục Tiêu
Đặc tả chi tiết hành vi người dùng, máy trạng thái, tiêu chí nghiệm thu và các trạng thái giao diện UI của tính năng Tạo phòng ban (`UC-01`).

---

# 1. Tổng Quan Tính Năng
Cho phép Quản trị viên nhập thông tin để tạo mới một phòng ban trên sơ đồ tổ chức.

---

# 2. User Stories
> Là một **Quản trị viên nhân sự**,  
> Tôi muốn **tạo một phòng ban mới trực thuộc Khối Công nghệ**,  
> Để **phản ánh đúng cơ cấu tổ chức mới của công ty và chuẩn bị tiếp nhận nhân sự**.

---

# 3. Hành Vi Chi Tiết & Thuật Toán Sinh Path

### Thuật toán sinh trường `path`:
```text
IF parentId IS NULL OR parentId == "" THEN
    newPath = "/" + newId + "/"
    depth = 1
ELSE
    parent = GetDepartment(parentId)
    ASSERT(parent.status == "ACTIVE")
    depth = countSegments(parent.path) + 1
    ASSERT(depth <= 5)
    newPath = parent.path + newId + "/"
END IF
```

- Ví dụ:
  - Phòng Gốc (`id = "root"`): `path = "/root/"` (Cấp 1)
  - Khối Công nghệ (`id = "tech"`, cha = "root"): `path = "/root/tech/"` (Cấp 2)
  - Phòng Backend (`id = "be"`, cha = "tech"): `path = "/root/tech/be/"` (Cấp 3)
  - Tổ API (`id = "api"`, cha = "be"): `path = "/root/tech/be/api/"` (Cấp 4)
  - Nhóm Core (`id = "core"`, cha = "api"): `path = "/root/tech/be/api/core/"` (Cấp 5)
  - Nếu tạo phòng dưới "core": `depth = 6 > 5` -> **Từ chối ngay lập tức**.

---

# 4. Máy Trạng Thái (State Machine)
Khi tạo mới thành công, phòng ban bước vào trạng thái khởi đầu:
```
[*] ── Tạo mới (UC-01) ──► ACTIVE
```

---

# 5. Hành Vi UI/UX (Đủ 6 Trạng Thái Bắt Buộc)

| Trạng thái | Hành vi giao diện người dùng |
| :--- | :--- |
| **Đang tải (Loading)** | Nút bấm hiển thị spinner, các ô nhập form bị disable. |
| **Rỗng (Empty)** | Khi mở modal, các trường mã và tên trống; phòng cha mặc định là phòng đang được chọn trên cây. |
| **Lỗi (Error)** | Hiển thị alert đỏ dưới ô nhập bị lỗi (Ví dụ: *"Mã TECH_BE đã được sử dụng"*). |
| **Không có quyền (403)** | Ẩn nút *"Thêm phòng ban"* đối với Trưởng phòng và Nhân viên thường. Nếu truy cập thẳng bằng URL -> Chuyển về màn hình 403 Forbidden. |
| **Đang lưu (Saving)** | Nút *"Lưu"* chuyển sang trạng thái loading, ngăn chặn người dùng bấm đúp. Tự động đính kèm `Idempotency-Key`. |
| **Mất kết nối (Offline)** | Hiển thị thông báo *"Mất kết nối internet. Không thể lưu phòng ban lúc này"*, giữ nguyên dữ liệu đã gõ trên form. |

---

# 6. Tiêu Chí Nghiệm Thu (Acceptance Criteria)
- [ ] Tạo trùng `code` -> Hệ thống từ chối, thông báo chỉ rõ mã nào trùng.
- [ ] Tạo trùng `name` trong cùng phòng cha -> Hệ thống từ chối.
- [ ] Tạo trùng `name` ở khác phòng cha -> Cho phép tạo thành công.
- [ ] Tạo phòng ban ở cấp thứ 6 -> Hệ thống từ chối (`BR-PB-03`).
- [ ] Đường dẫn `path` chuẩn xác tuyệt đối, không thiếu dấu gạch chéo `/`.
- [ ] Có đúng một bản ghi `AuditLog` được ghi trong CSDL.
