"use client";

import React, { useState, useEffect } from "react";
import Image from "next/image";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeft,
  Building2,
  Calendar,
  Mail,
  Briefcase,
  ShieldCheck,
  ArrowRightLeft,
  UserX,
  AlertTriangle,
  Clock,
  History,
  CheckCircle2,
  X,
  Loader2,
  Edit2,
  Camera,
  Trash2,
} from "lucide-react";
import { api, Employee, Department, AuditLog, getStoredUser, getAuditActionLabel } from "@/lib/api";
import { useOnline } from "@/lib/use-online";

export default function EmployeeDetailPage() {
  const router = useRouter();
  const params = useParams();
  const id = params?.id as string;

  const [employee, setEmployee] = useState<Employee | null>(null);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Transfer modal (UC-03)
  const [isTransferOpen, setIsTransferOpen] = useState(false);
  const [newDeptId, setNewDeptId] = useState("");
  const [confirmManagerRemoval, setConfirmManagerRemoval] = useState(false);
  const [managerPromptNeeded, setManagerPromptNeeded] = useState(false);

  // Resign modal (UC-04)
  const [isResignOpen, setIsResignOpen] = useState(false);
  const [resignReason, setResignReason] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editName, setEditName] = useState("");
  const [editTitle, setEditTitle] = useState("");
  const [editJoinedAt, setEditJoinedAt] = useState("");
  const [editAvatar, setEditAvatar] = useState("");
  const [avatarProcessing, setAvatarProcessing] = useState(false);
  const [editError, setEditError] = useState<string | null>(null);

  const currentUser = getStoredUser();
  const isOnline = useOnline();

  const loadData = async () => {
    if (!id) return;
    setIsLoading(true);
    setError(null);
    try {
      const [emp, depts, logs] = await Promise.all([
        api.getEmployeeById(id),
        api.getDepartments(),
        currentUser?.role === "admin" ? api.getAuditLogs(id) : Promise.resolve([]),
      ]);
      setEmployee(emp);
      setDepartments(depts);
      setAuditLogs(logs || []);
    } catch (err: any) {
      setError(err.message || "Không tìm thấy nhân viên hoặc không có quyền truy cập (404)");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [id]);

  const handleTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      await api.transferEmployee(id, newDeptId, confirmManagerRemoval);
      setSuccessMsg("Chuyển phòng ban thành công!");
      setIsTransferOpen(false);
      setManagerPromptNeeded(false);
      setConfirmManagerRemoval(false);
      await loadData();
    } catch (err: any) {
      if (err.message && err.message.includes("BR-NV-05")) {
        setManagerPromptNeeded(true);
      } else {
        setError(err.message || "Lỗi chuyển phòng ban");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleResign = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      await api.resignEmployee(id, resignReason);
      setSuccessMsg("Cập nhật nhân viên thôi việc (xóa mềm) thành công!");
      setIsResignOpen(false);
      await loadData();
    } catch (err: any) {
      setError(err.message || "Lỗi cập nhật thôi việc");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateProfile = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!employee) return;
    setIsSubmitting(true);
    setEditError(null);
    try {
      await api.updateEmployee(employee.id, {
        fullName: editName, title: editTitle, joinedAt: editJoinedAt, version: employee.version,
        ...(editAvatar !== (employee.avatar || "") ? { avatar: editAvatar } : {}),
      });
      window.dispatchEvent(new Event("minihrm:profile-updated"));
      setIsEditOpen(false);
      setSuccessMsg("Đã cập nhật hồ sơ nhân viên.");
      await loadData();
    } catch (err: any) {
      setEditError(err instanceof Error ? err.message : "Không thể cập nhật hồ sơ");
    } finally { setIsSubmitting(false); }
  };

  const handleAvatarFile = async (file?: File) => {
    if (!file) return;
    setEditError(null);
    if (!["image/jpeg", "image/png", "image/webp"].includes(file.type) || file.size > 5 * 1024 * 1024) {
      setEditError("Chọn ảnh JPG, PNG hoặc WebP dưới 5 MB.");
      return;
    }
    setAvatarProcessing(true);
    try {
      const bitmap = await createImageBitmap(file);
      try {
        const canvas = document.createElement("canvas");
        const scale = Math.min(1, 320 / Math.max(bitmap.width, bitmap.height));
        canvas.width = Math.max(1, Math.round(bitmap.width * scale));
        canvas.height = Math.max(1, Math.round(bitmap.height * scale));
        const context = canvas.getContext("2d");
        if (!context) throw new Error("Trình duyệt không thể xử lý ảnh này.");
        context.fillStyle = "#ffffff";
        context.fillRect(0, 0, canvas.width, canvas.height);
        context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
        let photo = canvas.toDataURL("image/jpeg", 0.75);
        if (photo.length > 87_000) photo = canvas.toDataURL("image/jpeg", 0.55);
        if (photo.length > 87_000) {
          canvas.width = Math.max(1, Math.round(canvas.width * 0.75));
          canvas.height = Math.max(1, Math.round(canvas.height * 0.75));
          context.fillStyle = "#ffffff";
          context.fillRect(0, 0, canvas.width, canvas.height);
          context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
          photo = canvas.toDataURL("image/jpeg", 0.55);
        }
        if (photo.length > 87_000) throw new Error("Ảnh quá lớn sau khi thu nhỏ. Vui lòng chọn ảnh khác.");
        setEditAvatar(photo);
      } finally { bitmap.close(); }
    } catch (err) {
      setEditError(err instanceof Error ? err.message : "Không thể đọc tệp ảnh.");
    } finally { setAvatarProcessing(false); }
  };

  if (isLoading) {
    return (
      <div aria-label="Đang tải hồ sơ nhân viên" className="mx-auto max-w-5xl space-y-4">
        <div className="h-6 w-24 animate-pulse rounded bg-slate-100 dark:bg-slate-800" />
        <div className="space-y-5 rounded-2xl border border-slate-200 bg-white p-6 dark:border-slate-800 dark:bg-slate-900">
          <div className="flex items-center gap-4"><div className="h-16 w-16 animate-pulse rounded-2xl bg-slate-100 dark:bg-slate-800" /><div className="space-y-2"><div className="h-5 w-52 animate-pulse rounded bg-slate-100 dark:bg-slate-800" /><div className="h-3 w-36 animate-pulse rounded bg-slate-100 dark:bg-slate-800" /></div></div>
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">{[0, 1, 2, 3].map((item) => <div key={item} className="h-20 animate-pulse rounded-xl bg-slate-100 dark:bg-slate-800" />)}</div>
        </div>
      </div>
    );
  }

  if (error || !employee) {
    return (
      <div className="mx-auto max-w-4xl space-y-4">
        <button onClick={() => router.back()} className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-800">
          <ArrowLeft className="w-4 h-4" />
          <span>Quay lại</span>
        </button>
        <div className="space-y-2 rounded-2xl border border-rose-200 bg-white p-6 text-rose-800 shadow-sm dark:bg-slate-900">
          <h2 className="text-base font-bold">Không thể mở hồ sơ</h2>
          <p className="text-sm">{error || "Hồ sơ nhân viên không tồn tại hoặc nằm ngoài phạm vi được phân quyền."}</p>
        </div>
      </div>
    );
  }

  const isResigned = employee.status === "RESIGNED";

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <button onClick={() => router.back()} className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-800">
        <ArrowLeft className="w-4 h-4" />
        <span>Quay lại danh sách</span>
      </button>

      {/* Alerts */}
      {successMsg && (
        <div className="p-4 rounded-xl bg-emerald-50 text-emerald-700 border border-emerald-200 text-sm flex items-center justify-between">
          <span className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 shrink-0" />
            {successMsg}
          </span>
          <button onClick={() => setSuccessMsg(null)}><X className="w-4 h-4" /></button>
        </div>
      )}

      {/* Main Profile Card */}
      <div className="space-y-6 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm shadow-slate-200/40 dark:border-slate-800 dark:bg-slate-900 dark:shadow-none sm:p-6">
        <div className="flex min-w-0 items-center gap-4">
            {employee.avatar ? (
              <Image src={employee.avatar} alt={`Ảnh đại diện ${employee.fullName}`} width={64} height={64} unoptimized className="h-14 w-14 shrink-0 rounded-2xl object-cover sm:h-16 sm:w-16" />
            ) : (
              <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-indigo-600 text-xl font-bold text-white shadow-md sm:h-16 sm:w-16">{employee.fullName.charAt(0)}</div>
            )}
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                <h1 className="min-w-0 text-xl font-bold font-heading tracking-tight text-slate-900 dark:text-slate-100 sm:text-2xl">
                  {employee.fullName}
                </h1>
                <span className="inline-flex shrink-0 whitespace-nowrap rounded bg-slate-100 px-2.5 py-1 font-mono text-xs text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                  {employee.code}
                </span>
              </div>
              <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
                <p className="text-sm text-slate-500">{employee.title}</p>
                <span className="hidden text-slate-300 sm:inline">·</span>
                <span className={`inline-flex whitespace-nowrap rounded-md border px-2 py-0.5 text-xs font-medium ${isResigned ? "border-rose-200 bg-rose-50 text-rose-700" : "border-emerald-200 bg-emerald-50 text-emerald-700"}`}>
                  {isResigned ? "Đã nghỉ việc" : "Đang làm việc"}
                </span>
              </div>
            </div>
        </div>

        {(currentUser?.role === "admin" || currentUser?.role === "manager") && (
          <div className={`grid gap-2 border-t border-slate-100 pt-4 dark:border-slate-800 ${currentUser.role === "admin" && !isResigned ? "grid-cols-1 sm:grid-cols-2 lg:grid-cols-3" : "grid-cols-1 sm:grid-cols-2"}`}>
              <button onClick={() => { setEditName(employee.fullName); setEditTitle(employee.title); setEditJoinedAt(employee.joinedAt); setEditAvatar(employee.avatar || ""); setEditError(null); setIsEditOpen(true); }} className="inline-flex min-h-10 items-center justify-center gap-2 whitespace-nowrap rounded-lg border border-slate-200 px-3 py-2 text-xs font-medium text-slate-700 shadow-xs transition hover:bg-slate-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800">
                <Edit2 className="h-4 w-4 shrink-0" /><span>Chỉnh sửa hồ sơ</span>
              </button>
              {currentUser.role === "admin" && !isResigned && <>
              <button
                onClick={() => setIsTransferOpen(true)}
                disabled={!isOnline}
                className="inline-flex min-h-10 items-center justify-center gap-2 whitespace-nowrap rounded-lg border border-slate-200 px-3 py-2 text-xs font-medium text-slate-700 shadow-xs transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
              >
                <ArrowRightLeft className="h-4 w-4 shrink-0" />
                <span>Chuyển phòng ban</span>
              </button>
              <button
                onClick={() => setIsResignOpen(true)}
                disabled={!isOnline}
                className="inline-flex min-h-10 items-center justify-center gap-2 whitespace-nowrap rounded-lg bg-rose-600 px-3 py-2 text-xs font-medium text-white shadow-sm transition hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <UserX className="h-4 w-4 shrink-0" />
                <span>Cho nghỉ việc</span>
              </button>
              </>}
          </div>
        )}

        {/* Details Grid */}
        <div className="grid grid-cols-1 gap-4 rounded-xl border border-slate-100 bg-slate-50 p-4 text-xs dark:border-slate-800/60 dark:bg-slate-950/50 sm:grid-cols-2 xl:grid-cols-4">
          <div>
            <span className="text-slate-400">Phòng ban:</span>
            <div className="font-semibold text-slate-800 dark:text-slate-200 mt-0.5">{departments.find((department) => department.id === employee.departmentId)?.name || employee.departmentName || "Chưa gán phòng ban"}</div>
          </div>
          <div>
            <span className="text-slate-400">Email:</span>
            <div className="font-semibold text-slate-800 dark:text-slate-200 mt-0.5">{employee.email}</div>
          </div>
          <div>
            <span className="text-slate-400">Ngày vào làm:</span>
            <div className="mt-0.5 font-mono tabular-nums text-slate-800 dark:text-slate-200">{employee.joinedAt}</div>
          </div>
          <div>
            <span className="text-slate-400">Phiên bản OCC:</span>
            <div className="font-mono text-slate-800 dark:text-slate-200 mt-0.5">v{employee.version}</div>
          </div>
        </div>

        {isResigned && (
          <div className="p-4 bg-amber-50 border border-amber-200 rounded-xl text-xs text-amber-800 space-y-1">
            <span className="font-semibold">Thông tin nghỉ việc (Xóa mềm - Ràng buộc 4):</span>
            <div>Ngày chính thức nghỉ: <strong>{employee.resignedAt}</strong></div>
            <div>Lý do: <em>{employee.resignReason}</em></div>
          </div>
        )}
      </div>

      {isEditOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4">
          <form onSubmit={handleUpdateProfile} className="max-h-[90dvh] w-full max-w-lg space-y-4 overflow-y-auto rounded-2xl bg-white p-5 shadow-xl dark:bg-slate-900">
            <div className="flex items-center justify-between"><h2 className="font-semibold">Sửa hồ sơ nhân viên</h2><button type="button" onClick={() => setIsEditOpen(false)}><X className="h-5 w-5" /></button></div>
            <div className="flex flex-wrap items-center gap-4 rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950/50">
              {editAvatar ? <Image src={editAvatar} alt="Xem trước ảnh đại diện" width={80} height={80} unoptimized className="h-20 w-20 shrink-0 rounded-xl object-cover" /> : <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-xl bg-indigo-100 text-2xl font-semibold text-indigo-700">{editName.charAt(0) || "?"}</div>}
              <div className="min-w-0 flex-1 space-y-2">
                <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Ảnh đại diện</p>
                <p className="text-xs text-slate-500">Ảnh JPG, PNG hoặc WebP. Ảnh được thu nhỏ trước khi lưu.</p>
                <div className="flex flex-wrap gap-2">
                  <label className="inline-flex cursor-pointer items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs font-medium text-indigo-700 hover:bg-indigo-50 focus-within:ring-2 focus-within:ring-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-indigo-300">
                    <Camera className="h-4 w-4" /> {editAvatar ? "Thay ảnh" : "Thêm ảnh"}
                    <input type="file" accept="image/jpeg,image/png,image/webp" disabled={avatarProcessing || isSubmitting || !isOnline} onChange={(event) => { void handleAvatarFile(event.target.files?.[0]); event.target.value = ""; }} className="sr-only" />
                  </label>
                  {editAvatar && <button type="button" onClick={() => setEditAvatar("")} disabled={avatarProcessing || isSubmitting || !isOnline} className="inline-flex items-center gap-1.5 rounded-lg border border-rose-200 px-3 py-2 text-xs font-medium text-rose-700 hover:bg-rose-50"><Trash2 className="h-4 w-4" /> Xóa ảnh</button>}
                </div>
              </div>
            </div>
            {editError && <p role="alert" className="rounded-lg border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700">{editError}</p>}
            <label className="block text-xs font-medium">Họ và tên<input required minLength={2} maxLength={100} value={editName} onChange={(event) => setEditName(event.target.value)} className="mt-1 h-10 w-full rounded-lg border px-3 text-sm dark:bg-slate-950" /></label>
            <label className="block text-xs font-medium">Chức danh<input required value={editTitle} onChange={(event) => setEditTitle(event.target.value)} className="mt-1 h-10 w-full rounded-lg border px-3 text-sm dark:bg-slate-950" /></label>
            <label className="block text-xs font-medium">Ngày vào làm<input required type="date" value={editJoinedAt} onChange={(event) => setEditJoinedAt(event.target.value)} className="mt-1 h-10 w-full rounded-lg border px-3 text-sm dark:bg-slate-950" /></label>
            <div className="flex justify-end gap-2"><button type="button" onClick={() => setIsEditOpen(false)} className="rounded-lg border px-3 py-2 text-sm">Hủy</button><button disabled={isSubmitting || avatarProcessing || !isOnline} className="rounded-lg bg-indigo-600 px-3 py-2 text-sm text-white disabled:opacity-50">{avatarProcessing ? "Đang xử lý ảnh…" : isSubmitting ? "Đang lưu…" : "Lưu thay đổi"}</button></div>
          </form>
        </div>
      )}

      {/* Audit Trail for this Employee */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 shadow-xs space-y-4">
        <div className="flex items-center gap-2">
          <History className="w-5 h-5 text-indigo-600" />
          <h2 className="text-base font-bold text-slate-900 dark:text-slate-100">
            Lịch Sử Kiểm Toán (Audit Log - Ràng Buộc 3)
          </h2>
        </div>

        {auditLogs.length === 0 ? (
          <p className="text-xs text-slate-500">Chưa ghi nhận sự kiện biến động nào.</p>
        ) : (
          <div className="space-y-3">
            {auditLogs.map(log => (
              <div key={log.id} className="p-3 rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50/50 text-xs flex items-start justify-between">
                <div>
                  <div className="font-semibold text-slate-800 dark:text-slate-200">
                    {getAuditActionLabel(log.action)} bởi {log.actorName || "người dùng hệ thống"}
                  </div>
                  <div className="text-slate-500 font-mono text-[11px] mt-0.5">
                    Thời gian: {new Date(log.occurredAt).toLocaleString("vi-VN", { dateStyle: "medium", timeStyle: "short" })}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* MODAL UC-03: TRANSFER */}
      {isTransferOpen && (
        <div className="fixed inset-y-0 inset-x-0 z-50 flex items-center justify-center bg-slate-950/40 backdrop-blur-xs p-4">
          <div className="w-full max-w-md bg-white dark:bg-slate-900 border border-slate-200 rounded-2xl shadow-xl overflow-hidden p-5 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 pb-3">
              <h3 className="font-bold text-sm">Chuyển Phòng Ban (UC-03)</h3>
              <button onClick={() => setIsTransferOpen(false)}><X className="w-4 h-4" /></button>
            </div>

            <form onSubmit={handleTransfer} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold mb-1">Chọn phòng ban đích (ACTIVE)</label>
                <select
                  required
                  value={newDeptId}
                  onChange={(e) => setNewDeptId(e.target.value)}
                  className="w-full h-9 px-2 rounded-lg border text-sm"
                >
                  <option value="">-- Chọn phòng ban --</option>
                  {departments.filter(d => d.id !== employee.departmentId && d.status === 'ACTIVE').map(d => (
                    <option key={d.id} value={d.id}>{d.name}</option>
                  ))}
                </select>
              </div>

              {managerPromptNeeded && (
                <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg text-xs text-amber-800 space-y-2">
                  <p className="font-semibold flex items-center gap-1">
                    <AlertTriangle className="w-4 h-4 shrink-0" />
                    <span>Quy tắc BR-NV-05: Chuyển Trưởng Phòng</span>
                  </p>
                  <p>Nhân sự này hiện đang là Trưởng phòng của phòng ban cũ. Bạn có đồng ý gỡ chức vụ Trưởng phòng cũ để chuyển đi?</p>
                  <label className="flex items-center gap-2 cursor-pointer font-medium">
                    <input
                      type="checkbox"
                      checked={confirmManagerRemoval}
                      onChange={(e) => setConfirmManagerRemoval(e.target.checked)}
                      className="rounded"
                    />
                    <span>Tôi xác nhận đồng ý gỡ chức vụ Trưởng phòng cũ</span>
                  </label>
                </div>
              )}

              <div className="flex justify-end gap-2 pt-2 border-t">
                <button type="button" onClick={() => setIsTransferOpen(false)} className="px-3 py-1.5 text-xs border rounded-lg">Hủy</button>
                <button type="submit" disabled={isSubmitting || !isOnline} className="px-4 py-1.5 text-xs bg-indigo-600 text-white rounded-lg font-medium flex items-center gap-1">
                  {isSubmitting && <Loader2 className="w-3 h-3 animate-spin" />}
                  <span>Xác nhận chuyển</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL UC-04: RESIGN */}
      {isResignOpen && (
        <div className="fixed inset-y-0 inset-x-0 z-50 flex items-center justify-center bg-slate-950/40 backdrop-blur-xs p-4">
          <div className="w-full max-w-md bg-white dark:bg-slate-900 border border-slate-200 rounded-2xl shadow-xl overflow-hidden p-5 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 pb-3">
              <h3 className="font-bold text-sm text-rose-700">Cho Thôi Việc Nhân Sự (UC-04)</h3>
              <button onClick={() => setIsResignOpen(false)}><X className="w-4 h-4" /></button>
            </div>

            <form onSubmit={handleResign} className="space-y-4">
              <div className="p-3 bg-rose-50 border border-rose-200 rounded-lg text-xs text-rose-800">
                Quy tắc <strong>BR-NV-04</strong>: Hệ thống sẽ chặn nếu nhân sự đang là Trưởng phòng của bất kỳ phòng ban ACTIVE nào chưa bàn giao chức vụ.
              </div>

              <div>
                <label className="block text-xs font-semibold mb-1">Lý do thôi việc</label>
                <textarea
                  rows={3}
                  value={resignReason}
                  onChange={(e) => setResignReason(e.target.value)}
                  placeholder="Ví dụ: Theo nguyện vọng cá nhân..."
                  className="w-full p-2.5 rounded-lg border text-sm"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t">
                <button type="button" onClick={() => setIsResignOpen(false)} className="px-3 py-1.5 text-xs border rounded-lg">Hủy</button>
                <button type="submit" disabled={isSubmitting || !isOnline} className="px-4 py-1.5 text-xs bg-rose-600 text-white rounded-lg font-medium flex items-center gap-1">
                  {isSubmitting && <Loader2 className="w-3 h-3 animate-spin" />}
                  <span>Xác nhận cho nghỉ việc</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
