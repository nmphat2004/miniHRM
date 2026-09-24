"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Users, UserPlus, Search, Filter, Mail, Building2, Briefcase, Calendar, ChevronRight, Loader2, RefreshCw, X, CheckCircle2 } from "lucide-react";
import { api, apiErrorMessage, Employee, Department, getStoredUser } from "@/lib/api";
import { useOnline } from "@/lib/use-online";

export default function EmployeesPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [search, setSearch] = useState("");
  const [selectedDept, setSelectedDept] = useState("");
  const [selectedStatus, setSelectedStatus] = useState("ACTIVE");
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Modal create employee
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [code, setCode] = useState("");
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [deptId, setDeptId] = useState("");
  const [title, setTitle] = useState("");
  const [joinedAt, setJoinedAt] = useState(new Date().toISOString().split('T')[0]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const currentUser = getStoredUser();
  const isOnline = useOnline();

  const loadData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [emps, depts] = await Promise.all([
        api.getEmployees(selectedDept, selectedStatus, search),
        api.getDepartments(),
      ]);
      setEmployees(emps || []);
      setDepartments(depts || []);
    } catch (err: any) {
      setError(apiErrorMessage(err, "Lỗi tải danh sách nhân sự"));
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedDept, selectedStatus, search]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      const idemKey = 'emp-create-' + Date.now();
      await api.createEmployee({
        code,
        fullName,
        email,
        departmentId: deptId,
        title,
        joinedAt,
      }, idemKey);
      setSuccessMsg(`Đã thêm nhân viên "${fullName}" thành công!`);
      setIsCreateOpen(false);
      setCode("");
      setFullName("");
      setEmail("");
      setTitle("");
      await loadData();
    } catch (err: any) {
      setError(apiErrorMessage(err, "Lỗi thêm nhân viên"));
    } finally {
      setIsSubmitting(false);
    }
  };

  const filteredEmployees = employees.filter(e =>
    e.fullName.toLowerCase().includes(search.toLowerCase()) ||
    e.code.toLowerCase().includes(search.toLowerCase()) ||
    e.email.toLowerCase().includes(search.toLowerCase()) ||
    e.title.toLowerCase().includes(search.toLowerCase())
  );
  const pageCount = Math.max(1, Math.ceil(filteredEmployees.length / pageSize));
  const visibleEmployees = filteredEmployees.slice((page - 1) * pageSize, page * pageSize);
  const departmentNameById = new Map(departments.map((department) => [department.id, department.name]));

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Users className="w-6 h-6 text-indigo-600" />
            <h1 className="text-2xl font-bold font-heading tracking-tight text-slate-900 dark:text-slate-100">
              Danh sách nhân sự
            </h1>
          </div>
          <p className="text-sm text-slate-500 mt-1">
            Quản lý hồ sơ nhân viên, chuyển phòng ban (UC-03) và thôi việc xóa mềm (UC-04).
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <button onClick={loadData} className="p-2 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-600 hover:bg-slate-50">
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
          {currentUser?.role === 'admin' && (
            <button
              onClick={() => setIsCreateOpen(true)}
              disabled={!isOnline}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-700 shadow-sm"
            >
              <UserPlus className="w-4 h-4" />
              <span>Thêm Nhân Viên Mới</span>
            </button>
          )}
        </div>
      </div>

      {/* Alerts */}
      {error && (
        <div className="p-4 rounded-xl bg-rose-50 text-rose-700 border border-rose-200 text-sm flex items-center justify-between">
          <span>{error}</span>
          <button onClick={() => setError(null)}><X className="w-4 h-4" /></button>
        </div>
      )}
      {successMsg && (
        <div className="p-4 rounded-xl bg-emerald-50 text-emerald-700 border border-emerald-200 text-sm flex items-center justify-between">
          <span className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 shrink-0" />
            {successMsg}
          </span>
          <button onClick={() => setSuccessMsg(null)}><X className="w-4 h-4" /></button>
        </div>
      )}

      {/* Filter toolbar */}
      <div className="flex flex-wrap items-center gap-3 p-4 bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-xl shadow-xs">
        <div className="relative flex-1 min-w-[240px]">
          <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-400" />
          <input
            type="text"
            placeholder="Tìm theo họ tên, mã nhân viên hoặc email..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="w-full h-9 pl-9 pr-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-slate-50 dark:bg-slate-950"
          />
        </div>

        <select
          value={selectedDept}
          onChange={(e) => { setSelectedDept(e.target.value); setPage(1); }}
          className="h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300"
        >
          <option value="">-- Tất cả phòng ban --</option>
          {departments.map(d => (
            <option key={d.id} value={d.id}>{d.name}</option>
          ))}
        </select>

        <select
          value={selectedStatus}
          onChange={(e) => { setSelectedStatus(e.target.value); setPage(1); }}
          className="h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300"
        >
          <option value="ALL">-- Tất cả trạng thái --</option>
          <option value="ACTIVE">Đang làm việc (ACTIVE)</option>
          <option value="RESIGNED">Đã nghỉ việc (RESIGNED)</option>
        </select>
      </div>

      {/* Employees Table */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl shadow-sm shadow-slate-200/40 dark:shadow-none overflow-hidden">
        {isLoading ? (
          <div aria-label="Đang tải danh sách nhân viên" className="divide-y divide-slate-100 p-5 dark:divide-slate-800">
            {[0, 1, 2, 3, 4].map((row) => <div key={row} className="flex items-center gap-4 py-4"><div className="h-10 w-10 animate-pulse rounded-xl bg-slate-100 dark:bg-slate-800" /><div className="flex-1 space-y-2"><div className="h-3 w-40 animate-pulse rounded bg-slate-100 dark:bg-slate-800" /><div className="h-2.5 w-64 max-w-full animate-pulse rounded bg-slate-100 dark:bg-slate-800" /></div><div className="hidden h-3 w-24 animate-pulse rounded bg-slate-100 dark:bg-slate-800 sm:block" /></div>)}
          </div>
        ) : filteredEmployees.length === 0 ? (
          <div className="flex flex-col items-center py-14 text-center">
            <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-indigo-50 text-indigo-600 dark:bg-indigo-950/60 dark:text-indigo-300"><Users className="h-5 w-5" /></div>
            <p className="text-sm font-semibold text-slate-800 dark:text-slate-200">Không tìm thấy nhân viên</p>
            <p className="mt-1 text-xs text-slate-500">Thử thay đổi từ khóa hoặc điều kiện lọc.</p>
          </div>
        ) : (
          <div className="overflow-x-auto"><table className="w-full min-w-[1080px] table-fixed text-left text-sm">
            <colgroup>
              <col className="w-[25%]" /><col className="w-[23%]" /><col className="w-[18%]" />
              <col className="w-[12%]" /><col className="w-[12%]" /><col className="w-[10%]" />
            </colgroup>
            <thead className="border-b border-slate-200/80 bg-slate-50/80 text-xs font-semibold tracking-wide text-slate-500 dark:border-slate-800 dark:bg-slate-950/50">
              <tr>
                <th className="whitespace-nowrap p-4 pl-6">Nhân viên</th>
                <th className="whitespace-nowrap p-4">Phòng ban</th>
                <th className="whitespace-nowrap p-4">Chức danh</th>
                <th className="whitespace-nowrap p-4">Ngày vào làm</th>
                <th className="whitespace-nowrap p-4">Trạng thái</th>
                <th className="whitespace-nowrap p-4 pr-6 text-right">Hồ sơ</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {visibleEmployees.map((emp) => (
                <tr key={emp.id} className="hover:bg-slate-50/60 dark:hover:bg-slate-800/40 transition-colors">
                  <td className="p-4 pl-6">
                    <div className="font-semibold text-slate-900 dark:text-slate-100">{emp.fullName}</div>
                    <div className="mt-1 flex min-w-0 items-center gap-2 whitespace-nowrap font-mono text-xs text-slate-400">
                      <span className="shrink-0">{emp.code}</span>
                      <span aria-hidden="true">·</span>
                      <span className="min-w-0 truncate">{emp.email}</span>
                    </div>
                  </td>
                  <td className="p-4 text-slate-600 dark:text-slate-400">
                    <div className="font-medium text-slate-800 dark:text-slate-200">{departmentNameById.get(emp.departmentId) || emp.departmentName || "Chưa gán phòng ban"}</div>
                  </td>
                  <td className="p-4 text-slate-700 dark:text-slate-300 font-medium">
                    {emp.title}
                  </td>
                  <td className="whitespace-nowrap p-4 font-mono text-xs tabular-nums text-slate-500">
                    {emp.joinedAt}
                  </td>
                  <td className="p-4">
                    {emp.status === 'ACTIVE' ? (
                      <span className="whitespace-nowrap rounded-md border border-emerald-200/60 bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700">
                        Đang làm việc
                      </span>
                    ) : (
                      <span className="whitespace-nowrap rounded-md border border-slate-200/60 bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600">
                        Đã nghỉ ({emp.resignedAt})
                      </span>
                    )}
                  </td>
                  <td className="p-4 pr-6 text-right">
                    <Link
                      href={`/employees/${emp.id}`}
                      className="inline-flex items-center justify-end gap-1 whitespace-nowrap rounded-lg px-2 py-1.5 text-xs font-medium text-indigo-600 transition hover:bg-indigo-50 hover:text-indigo-700"
                    >
                      <span>Xem hồ sơ</span>
                      <ChevronRight className="w-3.5 h-3.5" />
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table></div>
        )}
      </div>
      {!isLoading && filteredEmployees.length > 0 && (
        <div className="flex items-center justify-between text-sm text-slate-500">
          <span>Hiển thị {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, filteredEmployees.length)} / {filteredEmployees.length} nhân viên</span>
          <div className="flex items-center gap-2">
            <button type="button" disabled={page <= 1} onClick={() => setPage((current) => Math.max(1, current - 1))} className="rounded-md border px-3 py-1.5 disabled:opacity-40">Trước</button>
            <span>Trang {page} / {pageCount}</span>
            <button type="button" disabled={page >= pageCount} onClick={() => setPage((current) => Math.min(pageCount, current + 1))} className="rounded-md border px-3 py-1.5 disabled:opacity-40">Tiếp</button>
          </div>
        </div>
      )}

      {/* MODAL: THÊM NHÂN VIÊN */}
      {isCreateOpen && (
        <div className="fixed inset-y-0 inset-x-0 z-50 flex items-center justify-center bg-slate-950/40 backdrop-blur-xs p-4">
          <div className="w-full max-w-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-xl overflow-hidden">
            <div className="flex items-center justify-between p-5 border-b border-slate-100 dark:border-slate-800">
              <h3 className="font-semibold text-base text-slate-900 dark:text-slate-100">
                Thêm Nhân Viên Mới
              </h3>
              <button onClick={() => setIsCreateOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreate} className="p-5 space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Mã nhân viên (EMP-XXX)
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="EMP-012"
                    value={code}
                    onChange={(e) => setCode(e.target.value.toUpperCase())}
                    className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm font-mono"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Họ và tên (2-100 ký tự)
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Nguyễn Văn A"
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                  Email công ty (Duy nhất toàn hệ thống)
                </label>
                <input
                  type="email"
                  required
                  placeholder="name@minihrm.local"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Phòng ban trực thuộc (ACTIVE)
                  </label>
                  <select
                    required
                    value={deptId}
                    onChange={(e) => setDeptId(e.target.value)}
                    className="w-full h-9 px-2 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900"
                  >
                    <option value="">-- Chọn phòng ban --</option>
                    {departments.filter(d => d.status === 'ACTIVE').map(d => (
                      <option key={d.id} value={d.id}>{d.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                    Chức danh chuyên môn
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Kỹ sư phần mềm"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                  Ngày vào làm (Không muộn hơn hôm nay 90 ngày)
                </label>
                <input
                  type="date"
                  required
                  value={joinedAt}
                  onChange={(e) => setJoinedAt(e.target.value)}
                  className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm font-mono"
                />
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="px-4 py-2 rounded-lg border border-slate-200 text-xs font-medium text-slate-700"
                >
                  Hủy bỏ
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || !isOnline}
                  className="px-4 py-2 rounded-lg bg-indigo-600 text-white text-xs font-medium hover:bg-indigo-700 flex items-center gap-1.5"
                >
                  {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                  <span>Lưu nhân viên</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
