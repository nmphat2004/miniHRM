# Đặc Tả Hành Vi Tính Năng (Feature Spec) — Di Chuyển Phòng Ban

## 1. Tổng Quan
Đặc tả chi tiết hành vi, các trường hợp biên và tiêu chí nghiệm thu của Use Case `UC-02`.

---

# 2. User Stories
> Là một **Quản trị viên**,  
> Tôi muốn **chuyển Phòng Backend từ Khối Vận hành sang Khối Công nghệ**,  
> Để **tái sắp xếp nhân sự đúng theo định hướng công nghệ mới của công ty**.

---

# 3. Kịch Bản Kiểm Tra Nghiệp Vụ

```
       [Gốc]
      /     \
   [Tech]   [Ops]
             |
          [Backend] ◄── Muốn chuyển sang con của Tech
             |
           [API]
```

- **Kịch bản 1 (Di chuyển vào chính nó)**:
  - Chọn `sourceId = "Backend"`, `newParentId = "Backend"`.
  - Kết quả: Từ chối với lỗi *"Không thể di chuyển phòng ban vào chính nó"*.
- **Kịch bản 2 (Di chuyển vào con của nó)**:
  - Chọn `sourceId = "Backend"`, `newParentId = "API"`.
  - Kết quả: Từ chối với lỗi *"Không thể di chuyển phòng ban vào cây con của chính nó (BR-PB-04)"*.
- **Kịch bản 3 (Vượt quá 5 cấp)**:
  - Giả sử `Tech` đã là cấp 4, `Backend` có 2 cấp con bên dưới (tổng 3 cấp).
  - Ghép `Backend` vào `Tech` sẽ tạo nhánh cấp: `4 + 1 + 2 = 7 > 5`.
  - Kết quả: Từ chối với lỗi *"Nhánh cây sau khi di chuyển sẽ đạt 7 cấp, vượt quá giới hạn 5 cấp (BR-PB-03)"*.

---

# 4. Tiêu Chí Nghiệm Thu (Acceptance Criteria)
- [ ] Di chuyển vào chính nó -> Bị từ chối.
- [ ] Di chuyển vào cây con của chính nó -> Bị từ chối.
- [ ] Cây con gồm 3 cấp, 12 phòng ban -> Cả 12 phòng ban đều có `path` chuẩn xác sau khi di chuyển.
- [ ] Nhánh sâu nhất vượt quá 5 cấp sau khi di chuyển -> Bị từ chối trước khi ghi bất kỳ dữ liệu nào.
- [ ] Phạm vi dữ liệu của các Trưởng phòng cấp dưới tự động cập nhật ngay lập tức.
- [ ] Cập nhật hỏng giữa chừng -> Không bao giờ để cây ở trạng thái nửa vời.
