# Hệ Thống Thiết Kế — Design System (Mini HRM)

## 1. Triết Lý Thiết Kế (Design Philosophy)

Hệ thống Mini HRM là giải pháp quản trị doanh nghiệp hiện đại (Enterprise SaaS) dành cho 200–500 nhân sự. Thiết kế giao diện tập trung vào 3 giá trị cốt lõi:
1. **Rõ ràng & Giảm tải nhận thức (Clarity & Low Cognitive Load)**: Cấu trúc phân cấp nhiều tầng (Khối -> Phòng -> Tổ) phải được trực quan hóa mạch lạc, giúp người dùng nhận diện ngay vị trí của từng đơn vị và nhân sự mà không bị rối mắt.
2. **Chuẩn mực & Hiện đại (Enterprise Grade & Consistency)**: Xây dựng trên nền tảng **Tailwind CSS** + **shadcn/ui** (Radix UI primitives). Phong cách thiết kế phẳng, viền mảnh sắc sảo, đổ bóng mềm mại, hỗ trợ mượt mà cả **Light Mode** và **Dark Mode**.
3. **Phản hồi trạng thái trọn vẹn (6 Mandatory UI States)**: Không để người dùng hoang mang trong bất kỳ kịch bản mạng hay dữ liệu nào (Đang tải, Rỗng, Lỗi, Không có quyền, Đang lưu, Mất kết nối).

---

## 2. Bảng Màu Chuẩn (Color Palette & Design Tokens)

Mini HRM sử dụng hệ màu **Slate / Indigo / Emerald** chuẩn hóa theo Tailwind CSS:

### 2.1. Màu Thương Hiệu & Điểm Nhấn (Brand & Accent)
| Tên Token | Mã Hex (Light) | Lớp Tailwind | Ứng dụng nghiệp vụ |
| :--- | :--- | :--- | :--- |
| **Primary** | `#4F46E5` | `bg-indigo-600 text-white` | Nút hành động chính (CTA), điểm nhấn chọn, icon active |
| **Primary Hover** | `#4338CA` | `hover:bg-indigo-700` | Hiệu ứng rê chuột trên nút chính |
| **Primary Light** | `#EEF2FF` | `bg-indigo-50 text-indigo-700` | Nền badge cấp bậc, vùng chọn sáng |
| **Primary Dark** | `#312E81` | `dark:bg-indigo-950 text-indigo-300` | Nền điểm nhấn trong chế độ ban đêm |

### 2.2. Thang Màu Trung Tính (Neutral Scale - Slate)
| Token | Light Mode | Dark Mode | Mục đích sử dụng |
| :--- | :--- | :--- | :--- |
| **Background** | `#F8FAFC` (`slate-50`) | `#020617` (`slate-950`) | Nền tổng thể toàn trang dashboard |
| **Surface (Card)** | `#FFFFFF` (`white`) | `#0F172A` (`slate-900`) | Nền thẻ phòng ban, bảng dữ liệu, hộp thoại modal |
| **Border** | `#E2E8F0` (`slate-200`) | `#1E293B` (`slate-800`) | Đường viền ngăn cách, viền input, viền card |
| **Text Primary** | `#0F172A` (`slate-900`) | `#F8FAFC` (`slate-50`) | Tiêu đề chính, văn bản quan trọng, tên nhân viên |
| **Text Secondary** | `#64748B` (`slate-500`) | `#94A3B8` (`slate-400`) | Nhãn phụ, mã nhân viên, breadcrumbs, ngày tháng |
| **Text Muted** | `#94A3B8` (`slate-400`) | `#64748B` (`slate-500`) | Chữ giữ chỗ (Placeholder), ghi chú mờ |

### 2.3. Màu Trạng Thái Nghiệp Vụ (Semantic & Statuses)
| Trạng thái nghiệp vụ | Màu chủ đạo | Lớp Badge Tailwind | Ý nghĩa áp dụng |
| :--- | :--- | :--- | :--- |
| **Hoạt động (`ACTIVE`)** | **Emerald** | `bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/50 dark:text-emerald-300` | Phòng ban / Nhân sự đang làm việc bình thường |
| **Lưu trữ (`ARCHIVED`)** | **Slate / Muted** | `bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-800 dark:text-slate-300` | Phòng ban đã giải thể / ngừng hoạt động |
| **Nghỉ việc (`RESIGNED`)** | **Zinc / Gray** | `bg-zinc-100 text-zinc-600 border-zinc-200 dark:bg-zinc-800 dark:text-zinc-400` | Nhân sự đã thôi việc (Xóa mềm, khóa email) |
| **Cảnh báo / Đang chuyển** | **Amber** | `bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/50 dark:text-amber-300` | Cảnh báo Trưởng phòng cũ, trạng thái `MOVING` |
| **Lỗi / Xung đột OCC** | **Rose / Red** | `bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/50 dark:text-rose-300` | Lỗi form, xung đột 409, thao tác bị từ chối |

### 2.4. Mã Màu Phân Cấp Cây Phòng Ban (Depth Level Badges - Cấp 1 đến 5)
Để trực quan hóa ngay cấp độ của phòng ban trên sơ đồ cây:
- **Cấp 1 (Gốc - Công ty)**: `bg-indigo-100 text-indigo-800 border-indigo-300` (Ví dụ: Tập đoàn / Tổng công ty)
- **Cấp 2 (Khối)**: `bg-blue-100 text-blue-800 border-blue-300` (Ví dụ: Khối Công nghệ, Khối Kinh doanh)
- **Cấp 3 (Phòng)**: `bg-teal-100 text-teal-800 border-teal-300` (Ví dụ: Phòng Backend, Phòng Marketing)
- **Cấp 4 (Tổ)**: `bg-violet-100 text-violet-800 border-violet-300` (Ví dụ: Tổ API Platform, Tổ Sales B2B)
- **Cấp 5 (Nhóm)**: `bg-purple-100 text-purple-800 border-purple-300` (Cấp lá tối đa cho phép - `BR-PB-03`)

---

## 3. Kiểu Chữ (Typography)

- **Phông chữ chính (Primary Font)**: `Inter` hoặc `Plus Jakarta Sans`, sans-serif.
- **Phông mã lệnh / ID (Monospace)**: `JetBrains Mono` hoặc `Fira Code` dùng cho: Mã phòng ban (`code`), Mã nhân viên (`code`), UUID, chuỗi `path` (`/root/dept/`), giá trị Before/After diff.

### Thang Kích Thước (Type Scale)
| Cấp độ | Cỡ chữ / Line Height | Độ đậm (Weight) | Lớp Tailwind | Áp dụng |
| :--- | :--- | :--- | :--- | :--- |
| **Display / H1** | 28px / 36px | Bold (700) | `text-2xl font-bold tracking-tight` | Tiêu đề chính màn hình |
| **Heading / H2** | 20px / 28px | SemiBold (600) | `text-xl font-semibold` | Tiêu đề phân đoạn, Modal Title |
| **Subheading / H3** | 16px / 24px | SemiBold (600) | `text-base font-semibold` | Tên Node phòng ban trên cây |
| **Body Regular** | 14px / 20px | Normal (400) | `text-sm font-normal` | Nội dung bảng, hồ sơ, form label |
| **Body Medium** | 14px / 20px | Medium (500) | `text-sm font-medium` | Tên người dùng, nút bấm, tab active |
| **Caption / Small** | 12px / 16px | Normal (400) | `text-xs text-muted-foreground` | Breadcrumbs, thời gian, chú thích |
| **Mono Code** | 13px / 18px | Normal (400) | `font-mono text-xs` | `path`, mã nhân viên, `Idempotency-Key` |

---

## 4. Khoảng Cách & Bo Góc (Spacing & Elevation)

- **Lưới khoảng cách (Spacing Grid)**: Dựa trên bội số của 4px (`p-1 = 4px`, `p-2 = 8px`, `p-4 = 16px`, `p-6 = 24px`).
- **Bo góc (Border Radius)**:
  - Nút bấm, ô nhập liệu (Input/Button): `rounded-md` (6px).
  - Thẻ phòng ban, Card, Hộp thoại Modal: `rounded-xl` (12px) hoặc `rounded-lg` (8px).
  - Huy hiệu (Badge, Status Pill), Avatar: `rounded-full`.
- **Đổ bóng (Box Shadow)**:
  - Thẻ thông thường: `shadow-sm border border-border` (rất nhẹ và tinh tế).
  - Thẻ Node khi Hover hoặc Đang kéo: `shadow-md border-indigo-400`.
  - Hộp thoại Modal / Dropdown: `shadow-xl`.

---

## 5. Đặc Tả 6 Trạng Thái Giao Diện Bắt Buộc (The 6 UI States)

Theo **Ràng buộc số 8**, mọi màn hình danh sách và chi tiết bắt buộc phải hiện thực hóa đủ 6 trạng thái này:

```
  ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
  │ 1. ĐANG TẢI     │       │ 2. RỖNG         │       │ 3. LỖI          │
  │ Shimmer Skeleton│       │ Empty Illustration│     │ Error Box + Retry│
  └─────────────────┘       └─────────────────┘       └─────────────────┘
  ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
  │ 4. KHÔNG QUYỀN  │       │ 5. ĐANG LƯU     │       │ 6. MẤT KẾT NỐI  │
  │ 404 Shield / 403│       │ Button Spinner  │       │ Offline Banner  │
  └─────────────────┘       └─────────────────┘       └─────────────────┘
```

### 1. Trạng Thái Đang Tải (Loading State)
- **Cơ chế**: Tuyệt đối không dùng màn hình trắng trơn hoặc spinner đơn lẻ giữa trang.
- **Thực thi**: Sử dụng **Skeleton UI** (khung xương mô phỏng với hiệu ứng `animate-pulse` của Tailwind) dựng đúng cấu trúc bảng hoặc cây tổ chức.

### 2. Trạng Thái Rỗng (Empty State)
- **Cơ chế**: Khi một phòng ban chưa có phòng con, hoặc tìm kiếm nhân viên không có kết quả.
- **Thực thi**:
  - Icon minh họa trung tính (`Users`, `FolderOpen` mờ).
  - Thông điệp thân thiện: *"Chưa có nhân viên nào trong phòng ban này"*.
  - Nút hành động trực tiếp: `+ Thêm nhân viên` hoặc `Xóa bộ lọc`.

### 3. Trạng Thái Lỗi (Error State)
- **Cơ chế**: Khi API Go trả về lỗi 500 hoặc truy vấn CSDL thất bại.
- **Thực thi**:
  - Hộp thông báo màu đỏ nhạt (`bg-rose-50 border-rose-200`).
  - Hiển thị thông báo dễ hiểu kèm mã lỗi kỹ thuật phía dưới.
  - Nút bấm *"Thử lại (Retry)"* để kích hoạt nạp lại dữ liệu mà không cần F5 trình duyệt.

### 4. Trạng Thái Không Có Quyền (Forbidden / 404 Security State)
- **Cơ chế**: Áp dụng **Ràng buộc 2** khi người dùng xem hồ sơ/phòng ban ngoài phạm vi.
- **Thực thi**:
  - Hiển thị giao diện **`404 Not Found`** trung tính (*"Không tìm thấy hồ sơ hoặc bạn không có quyền truy cập"*).
  - Tuyệt đối không hiện `403` ở màn hình chi tiết để ngăn chặn kẻ tấn công dò quét ID hợp lệ.

### 5. Trạng Thái Đang Lưu (Saving / Submitting State)
- **Cơ chế**: Khi người dùng nhấn "Lưu", "Di chuyển phòng", hoặc "Chuyển nhân sự".
- **Thực thi**:
  - Nút bấm bị vô hiệu hóa (`disabled`), con trỏ chuột `cursor-not-allowed`.
  - Icon bên trong nút chuyển thành vòng tròn quay (`Loader2` animate-spin).
  - Tự động gắn `Idempotency-Key` vào request để ngăn chặn việc gửi nhiều lần.

### 6. Trạng Thái Mất Kết Nối (Offline State)
- **Cơ chế**: Khi trình duyệt phát hiện mất kết nối mạng (`window.navigator.onLine === false`) hoặc API Go timeout.
- **Thực thi**:
  - Thanh thông báo cố định trên cùng (Sticky Banner) màu vàng/hổ phách: *"Mất kết nối tới máy chủ. Vui lòng kiểm tra đường truyền"*.
  - Khóa các nút hành động ghi/sửa để tránh phát sinh lỗi không đáng có.

---

## 6. Thư Viện Thành Phần Cốt Lõi (Core Component Library)

### 6.1. Thẻ Phòng Ban Trên Cây Tổ Chức (Department Node Card)
```tsx
// Cấu trúc tiêu chuẩn cho Node phòng ban trên cây
<div className="group relative flex flex-col p-4 rounded-xl bg-card border border-border shadow-sm hover:border-indigo-400 hover:shadow-md transition-all duration-200">
  {/* Header: Mã phòng & Cấp bậc */}
  <div className="flex items-center justify-between gap-2 mb-2">
    <span className="font-mono text-xs font-semibold px-2 py-0.5 rounded bg-muted text-muted-foreground">
      {dept.code}
    </span>
    <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-indigo-50 text-indigo-700">
      Cấp {dept.depth}/5
    </span>
  </div>

  {/* Tên phòng */}
  <h4 className="text-sm font-semibold text-foreground mb-3 line-clamp-1">
    {dept.name}
  </h4>

  {/* Footer: Trưởng phòng & Nhân sự */}
  <div className="flex items-center justify-between text-xs text-muted-foreground pt-2 border-t border-border/60">
    <div className="flex items-center gap-1.5">
      <UserCircle className="w-3.5 h-3.5" />
      <span>{dept.managerName || "Khuyết Trưởng phòng"}</span>
    </div>
    <div className="flex items-center gap-1 font-medium">
      <Users className="w-3.5 h-3.5" />
      <span>{dept.employeeCount}</span>
    </div>
  </div>
</div>
```

### 6.2. Bộ So Sánh Biến Động Nhật Ký (Before / After Diff Viewer)
Dùng trong modal xem chi tiết nhật ký kiểm toán (`AuditLog`):
- **Trường Cũ (`before`)**: Nền đỏ nhạt (`bg-rose-50/70 text-rose-900 border-rose-200`), có icon gạch bỏ hoặc dấu trừ `-`.
- **Trường Mới (`after`)**: Nền xanh nhạt (`bg-emerald-50/70 text-emerald-900 border-emerald-200`), có icon dấu cộng `+`.
- Chỉ hiển thị các trường có sự thay đổi thật sự (ví dụ: `name`, `parentId`, `status`).

---

## 7. Cấu Hình CSS Variables Mẫu (`globals.css`)

```css
@layer base {
  :root {
    --background: 210 40% 98%;      /* slate-50 */
    --foreground: 222.2 84% 4.9%;   /* slate-900 */
    --card: 0 0% 100%;              /* white */
    --card-foreground: 222.2 84% 4.9%;
    --primary: 243.4 75.4% 58.6%;   /* indigo-600 */
    --primary-foreground: 210 40% 98%;
    --muted: 210 40% 96.1%;         /* slate-100 */
    --muted-foreground: 215.4 16.3% 46.9%; /* slate-500 */
    --border: 214.3 31.8% 91.4%;    /* slate-200 */
    --ring: 243.4 75.4% 58.6%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;   /* slate-950 */
    --foreground: 210 40% 98%;
    --card: 222.2 47.4% 11.2%;      /* slate-900 */
    --card-foreground: 210 40% 98%;
    --primary: 243.4 75.4% 58.6%;   /* indigo-500 */
    --primary-foreground: 222.2 47.4% 11.2%;
    --muted: 217.2 32.6% 17.5%;     /* slate-800 */
    --muted-foreground: 215 20.2% 65.1%;  /* slate-400 */
    --border: 217.2 32.6% 17.5%;
  }
}
```
