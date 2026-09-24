"use client";

import React, { useState, useEffect } from "react";
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
  Loader2
} from "lucide-react";
import { api, Employee, Department, AuditLog, getStoredUser } from "@/lib/api";

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

  const currentUser = getStoredUser();

  const loadData = async () => {
    if (!id) return;
    setIsLoading(true);
    setError(null);
    try {
      const [emp, depts, logs] = await Promise.all([
        api.getEmployeeById(id),
        api.getDepartments(),
        api.getAuditLogs(id),
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

  if (isLoading) {
    return (
      <div className="p-12 flex flex-col items-center justify-center text-slate-400 gap-2">
        <Loader2 className="w-8 h-8 animate-spin text-indigo-600" />
        <span className="text-sm">Đang tải hồ sơ nhân viên...</span>
      </div>
    );
  }

  if (error || !employee) {
    return (
      <div className="p-8 max-w-4xl mx-auto space-y-4">
        <button onClick={() => router.back()} className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-slate-800">
          <ArrowLeft className="w-4 h-4" />
          <span>Quay lại</span>
        </button>
        <div className="p-6 bg-rose-50 border border-rose-200 rounded-2xl text-rose-800 space-y-2">
          <h2 className="text-base font-bold">Lỗi truy cập dữ liệu (404 / Scoping)</h2>
          <p className="text-sm">{error || "Hồ sơ nhân viên không tồn tại hoặc nằm ngoài phạm vi được phân quyền."}</p>
        </div>
      </div>
    );
  }

  const isResigned = employee.status === "RESIGNED";

  return (
    <div className="p-8 max-w-5xl mx-auto space-y-6">
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
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 shadow-xs space-y-6">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-2xl bg-indigo-600 text-white font-bold text-xl flex items-center justify-center shadow-md">
              {employee.fullName.charAt(0)}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold font-heading text-slate-900 dark:text-slate-100">
                  {employee.fullName}
                </h1>
                <span className="px-2.5 py-0.5 rounded text-xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                  {employee.code}
                </span>
                {isResigned ? (
                  <span className="px-2.5 py-0.5 rounded text-xs font-medium bg-rose-50 text-rose-700 border border-rose-200">
                    Đã nghỉ việc (RESIGNED)
                  </span>
                ) : (
                  <span className="px-2.5 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-200">
                    Đang làm việc (ACTIVE)
                  </span>
                )}
              </div>
              <p className="text-sm text-slate-500 mt-1">{employee.title}</p>
            </div>
          </div>

          {currentUser?.role === 'admin' && !isResigned && (
            <div className="flex items-center gap-2">
              <button
                onClick={() => setIsTransferOpen(true)}
                className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-medium shadow-xs"
              >
                <ArrowRightLeft className="w-3.5 h-3.5" />
                <span>Chuyển Phòng Ban (UC-03)</span>
              </button>
              <button
                onClick={() => setIsResignOpen(true)}
                className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-rose-600 text-white hover:bg-rose-700 text-xs font-medium shadow-xs"
              >
                <UserX className="w-3.5 h-3.5" />
                <span>Cho Thôi Việc (UC-04)</span>
              </button>
            </div>
          )}
        </div>

        {/* Details Grid */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 rounded-xl bg-slate-50 dark:bg-slate-950/50 border border-slate-100 dark:border-slate-800/60 text-xs">
          <div>
            <span className="text-slate-400">Phòng ban:</span>
            <div className="font-semibold text-slate-800 dark:text-slate-200 mt-0.5">{employee.departmentName}</div>
            <div className="font-mono text-[10px] text-slate-400">{employee.departmentPath}</div>
          </div>
          <div>
            <span className="text-slate-400">Email:</span>
            <div className="font-semibold text-slate-800 dark:text-slate-200 mt-0.5">{employee.email}</div>
          </div>
          <div>
            <span className="text-slate-400">Ngày vào làm:</span>
            <div className="font-mono text-slate-800 dark:text-slate-200 mt-0.5">{employee.joinedAt}</div>
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
                    [{log.action}] thực hiện bởi {log.actorName}
                  </div>
                  <div className="text-slate-500 font-mono text-[11px] mt-0.5">
                    Thời gian: {log.occurredAt}
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
                <button type="submit" disabled={isSubmitting} className="px-4 py-1.5 text-xs bg-indigo-600 text-white rounded-lg font-medium flex items-center gap-1">
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
                <button type="submit" disabled={isSubmitting} className="px-4 py-1.5 text-xs bg-rose-600 text-white rounded-lg font-medium flex items-center gap-1">
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
