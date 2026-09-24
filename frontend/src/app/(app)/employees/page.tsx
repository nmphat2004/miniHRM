"use client";

import React, { useState, useEffect } from "react";
import Link from "next/navigation";
import { Users, UserPlus, Search, Filter, Mail, Building2, Briefcase, Calendar, ChevronRight, Loader2, RefreshCw, X, CheckCircle2 } from "lucide-react";
import { api, Employee, Department, getStoredUser } from "@/lib/api";

export default function EmployeesPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [search, setSearch] = useState("");
  const [selectedDept, setSelectedDept] = useState("");
  const [selectedStatus, setSelectedStatus] = useState("");
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

  const loadData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [emps, depts] = await Promise.all([
        api.getEmployees(selectedDept, selectedStatus),
        api.getDepartments(),
      ]);
      setEmployees(emps || []);
      setDepartments(depts || []);
    } catch (err: any) {
      setError(err.message || "Lỗi tải danh sách nhân sự");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedDept, selectedStatus]);

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
      setError(err.message || "Lỗi thêm nhân viên");
    } finally {
      setIsSubmitting(false);
    }
  };

  const filteredEmployees = employees.filter(e =>
    e.fullName.toLowerCase().includes(search.toLowerCase()) ||
    e.code.toLowerCase().includes(search.toLowerCase()) ||
    e.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Users className="w-6 h-6 text-indigo-600" />
            <h1 className="text-2xl font-bold font-heading text-slate-900 dark:text-slate-100">
              Danh Sách Nhân Sự
            </h1>
          </div>
          <p className="text-sm text-slate-500 mt-1">
            Quản lý hồ sơ nhân viên, chuyển phòng ban (UC-03) và thôi việc xóa mềm (UC-04).
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button onClick={loadData} className="p-2 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-600 hover:bg-slate-50">
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
          {currentUser?.role === 'admin' && (
            <button
              onClick={() => setIsCreateOpen(true)}
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
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-9 pl-9 pr-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-slate-50 dark:bg-slate-950"
          />
        </div>

        <select
          value={selectedDept}
          onChange={(e) => setSelectedDept(e.target.value)}
          className="h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300"
        >
          <option value="">-- Tất cả phòng ban --</option>
          {departments.map(d => (
            <option key={d.id} value={d.id}>{d.name}</option>
          ))}
        </select>

        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value)}
          className="h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300"
        >
          <option value="">-- Tất cả trạng thái --</option>
          <option value="ACTIVE">Đang làm việc (ACTIVE)</option>
          <option value="RESIGNED">Đã nghỉ việc (RESIGNED)</option>
        </select>
      </div>

      {/* Employees Table */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl shadow-xs overflow-hidden">
        {isLoading ? (
          <div className="py-12 flex flex-col items-center justify-center text-slate-400 gap-2">
            <Loader2 className="w-6 h-6 animate-spin text-indigo-600" />
            <span className="text-sm">Đang tải danh sách nhân viên...</span>
          </div>
        ) : filteredEmployees.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-sm">
            Không tìm thấy nhân viên nào phù hợp.
          </div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50/80 dark:bg-slate-950/50 border-b border-slate-200/80 dark:border-slate-800 text-xs font-semibold uppercase text-slate-500 tracking-wider">
              <tr>
                <th className="p-4 pl-6">Nhân Viên</th>
                <th className="p-4">Phòng Ban</th>
                <th className="p-4">Chức Danh</th>
                <th className="p-4">Ngày Vào Làm</th>
                <th className="p-4">Trạng Thái</th>
                <th className="p-4 pr-6 text-right">Chi Tiết</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {filteredEmployees.map((emp) => (
                <tr key={emp.id} className="hover:bg-slate-50/60 dark:hover:bg-slate-800/40 transition-colors">
                  <td className="p-4 pl-6">
                    <div className="font-semibold text-slate-900 dark:text-slate-100">{emp.fullName}</div>
                    <div className="text-xs text-slate-400 font-mono flex items-center gap-2 mt-0.5">
                      <span>{emp.code}</span>
                      <span>·</span>
                      <span>{emp.email}</span>
                    </div>
                  </td>
                  <td className="p-4 text-slate-600 dark:text-slate-400">
                    <div className="font-medium text-slate-800 dark:text-slate-200">{emp.departmentName || emp.departmentId}</div>
                    <div className="text-[11px] font-mono text-slate-400">{emp.departmentPath}</div>
                  </td>
                  <td className="p-4 text-slate-700 dark:text-slate-300 font-medium">
                    {emp.title}
                  </td>
                  <td className="p-4 text-slate-500 font-mono text-xs">
                    {emp.joinedAt}
                  </td>
                  <td className="p-4">
                    {emp.status === 'ACTIVE' ? (
                      <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-200/60">
                        Đang làm việc
                      </span>
                    ) : (
                      <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-slate-100 text-slate-600 border border-slate-200/60">
                        Đã nghỉ ({emp.resignedAt})
                      </span>
                    )}
                  </td>
                  <td className="p-4 pr-6 text-right">
                    <a
                      href={`/employees/${emp.id}`}
                      className="inline-flex items-center gap-1 text-xs font-medium text-indigo-600 hover:text-indigo-700 px-3 py-1.5 rounded-lg hover:bg-indigo-50"
                    >
                      <span>Xem hồ sơ</span>
                      <ChevronRight className="w-3.5 h-3.5" />
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

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
                  disabled={isSubmitting}
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
