"use client";

import React, { useState, useEffect } from "react";
import {
  Network,
  Plus,
  ArrowRightLeft,
  Archive,
  Edit2,
  Users,
  ChevronRight,
  ChevronDown,
  FolderTree,
  AlertTriangle,
  CheckCircle2,
  X,
  Loader2,
  UserCheck,
  RefreshCw
} from "lucide-react";
import { api, Department, getStoredUser } from "@/lib/api";

export default function DepartmentsPage() {
  const [departments, setDepartments] = useState<Department[]>([]);
  const [allFlatDepts, setAllFlatDepts] = useState<Department[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Expanded nodes state
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  // Modals state
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isMoveModalOpen, setIsMoveModalOpen] = useState(false);
  const [selectedDept, setSelectedDept] = useState<Department | null>(null);

  // Form states
  const [newCode, setNewCode] = useState("");
  const [newName, setNewName] = useState("");
  const [selectedParentId, setSelectedParentId] = useState("");
  const [moveTargetParentId, setMoveTargetParentId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const currentUser = getStoredUser();

  const loadData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [tree, flat] = await Promise.all([
        api.getDepartmentTree(),
        api.getDepartments(),
      ]);
      setDepartments(tree || []);
      setAllFlatDepts(flat || []);

      // Auto expand root nodes
      const expMap: Record<string, boolean> = {};
      (flat || []).forEach(d => { expMap[d.id] = true; });
      setExpanded(expMap);
    } catch (err: any) {
      setError(err.message || "Không thể tải cây phòng ban");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const toggleExpand = (id: string) => {
    setExpanded(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const handleCreateDepartment = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      const idemKey = 'dept-create-' + Date.now();
      await api.createDepartment(newCode, newName, selectedParentId, idemKey);
      setSuccessMsg(`Đã tạo phòng ban "${newName}" thành công!`);
      setIsCreateModalOpen(false);
      setNewCode("");
      setNewName("");
      setSelectedParentId("");
      await loadData();
    } catch (err: any) {
      setError(err.message || "Lỗi khi tạo phòng ban");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleExecuteMove = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedDept) return;
    setIsSubmitting(true);
    setError(null);
    try {
      await api.moveDepartment(selectedDept.id, moveTargetParentId);
      setSuccessMsg(`Đã di chuyển phòng ban "${selectedDept.name}" thành công!`);
      setIsMoveModalOpen(false);
      setSelectedDept(null);
      setMoveTargetParentId("");
      await loadData();
    } catch (err: any) {
      setError(err.message || "Lỗi di chuyển phòng ban");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleArchive = async (dept: Department) => {
    if (!confirm(`Bạn có chắc chắn muốn lưu trữ phòng ban "${dept.name}"?\nLưu ý: Không thể lưu trữ nếu còn nhân viên hoặc phòng ban con đang hoạt động (BR-PB-06).`)) {
      return;
    }
    setError(null);
    try {
      await api.archiveDepartment(dept.id);
      setSuccessMsg(`Đã lưu trữ phòng ban "${dept.name}" thành công!`);
      await loadData();
    } catch (err: any) {
      setError(err.message || "Lỗi lưu trữ phòng ban");
    }
  };

  const renderTree = (nodes: Department[], depth: number = 1) => {
    return nodes.map((dept) => {
      const isExp = expanded[dept.id] ?? true;
      const hasChildren = dept.children && dept.children.length > 0;
      const isArchived = dept.status === "ARCHIVED";

      return (
        <div key={dept.id} className="space-y-1">
          <div
            className={`group flex items-center justify-between p-3 rounded-xl border transition-all ${
              isArchived
                ? "bg-slate-50/60 dark:bg-slate-900/40 border-slate-200/60 dark:border-slate-800/60 opacity-70"
                : "bg-white dark:bg-slate-900 border-slate-200/80 dark:border-slate-800 hover:border-indigo-300 dark:hover:border-indigo-800 shadow-xs"
            }`}
            style={{ marginLeft: `${(depth - 1) * 24}px` }}
          >
            <div className="flex items-center gap-2.5 min-w-0">
              <button
                type="button"
                onClick={() => toggleExpand(dept.id)}
                className="w-5 h-5 flex items-center justify-center text-slate-400 hover:text-slate-600 rounded"
              >
                {hasChildren ? (
                  isExp ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />
                ) : (
                  <span className="w-1.5 h-1.5 rounded-full bg-slate-300" />
                )}
              </button>

              <div className="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 flex items-center justify-center shrink-0">
                <FolderTree className="w-4 h-4" />
              </div>

              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-sm text-slate-900 dark:text-slate-100 truncate">
                    {dept.name}
                  </span>
                  <span className="px-2 py-0.5 rounded text-[11px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                    {dept.code}
                  </span>
                  {isArchived && (
                    <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 border border-amber-200/60">
                      Đã lưu trữ (ARCHIVED)
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-3 text-xs text-slate-500 mt-0.5">
                  <span className="font-mono text-[11px] text-slate-400">{dept.path}</span>
                  {dept.managerName ? (
                    <span className="flex items-center gap-1 text-indigo-600 dark:text-indigo-400 font-medium">
                      <UserCheck className="w-3 h-3" />
                      {dept.managerName}
                    </span>
                  ) : (
                    <span className="text-slate-400 italic">Chưa có Trưởng phòng</span>
                  )}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <span className="flex items-center gap-1 text-xs text-slate-500 px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800">
                <Users className="w-3.5 h-3.5" />
                <strong>{dept.employeeCount}</strong> nhân sự
              </span>

              {currentUser?.role === 'admin' && (
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  {!isArchived && (
                    <button
                      onClick={() => {
                        setSelectedDept(dept);
                        setIsMoveModalOpen(true);
                      }}
                      title="Di chuyển phòng ban (UC-02)"
                      className="p-1.5 rounded-lg text-slate-500 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-950/50"
                    >
                      <ArrowRightLeft className="w-4 h-4" />
                    </button>
                  )}
                  {!isArchived && dept.employeeCount === 0 && (
                    <button
                      onClick={() => handleArchive(dept)}
                      title="Lưu trữ phòng ban (BR-PB-06)"
                      className="p-1.5 rounded-lg text-slate-500 hover:text-amber-600 hover:bg-amber-50 dark:hover:bg-amber-950/50"
                    >
                      <Archive className="w-4 h-4" />
                    </button>
                  )}
                </div>
              )}
            </div>
          </div>

          {hasChildren && isExp && (
            <div className="space-y-1">
              {renderTree(dept.children!, depth + 1)}
            </div>
          )}
        </div>
      );
    });
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Network className="w-6 h-6 text-indigo-600" />
            <h1 className="text-2xl font-bold font-heading text-slate-900 dark:text-slate-100">
              Sơ Đồ Cây Phòng Ban
            </h1>
          </div>
          <p className="text-sm text-slate-500 mt-1">
            Quản trị cấu trúc tổ chức phân cấp, di chuyển phòng ban (UC-02) và phân quyền Scoping (Ràng buộc 1).
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={loadData}
            className="p-2 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-600 hover:bg-slate-50"
            title="Tải lại dữ liệu"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
          {currentUser?.role === 'admin' && (
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-700 shadow-sm transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Thêm Phòng Ban (UC-01)</span>
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

      {/* Tree container */}
      <div className="bg-slate-50/50 dark:bg-slate-950/40 p-6 rounded-2xl border border-slate-200/80 dark:border-slate-800">
        {isLoading ? (
          <div className="py-12 flex flex-col items-center justify-center text-slate-400 gap-2">
            <Loader2 className="w-6 h-6 animate-spin text-indigo-600" />
            <span className="text-sm">Đang tải sơ đồ cây tổ chức...</span>
          </div>
        ) : departments.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-sm">
            Chưa có phòng ban nào trong phạm vi quản lý của bạn.
          </div>
        ) : (
          <div className="space-y-2">
            {renderTree(departments)}
          </div>
        )}
      </div>

      {/* MODAL UC-01: TẠO PHÒNG BAN */}
      {isCreateModalOpen && (
        <div className="fixed inset-y-0 inset-x-0 z-50 flex items-center justify-center bg-slate-950/40 backdrop-blur-xs p-4">
          <div className="w-full max-w-md bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-xl overflow-hidden">
            <div className="flex items-center justify-between p-5 border-b border-slate-100 dark:border-slate-800">
              <h3 className="font-semibold text-base text-slate-900 dark:text-slate-100">
                Thêm Phòng Ban Mới (UC-01)
              </h3>
              <button onClick={() => setIsCreateModalOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateDepartment} className="p-5 space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                  Mã phòng ban (A-Z0-9_, duy nhất toàn hệ thống)
                </label>
                <input
                  type="text"
                  placeholder="Ví dụ: TECH_MOBILE"
                  value={newCode}
                  onChange={(e) => setNewCode(e.target.value.toUpperCase())}
                  required
                  className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm font-mono"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                  Tên phòng ban (2-100 ký tự)
                </label>
                <input
                  type="text"
                  placeholder="Ví dụ: Tổ Phát Triển Mobile"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  required
                  className="w-full h-9 px-3 rounded-lg border border-slate-200 dark:border-slate-800 text-sm"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                  Phòng ban cha trực thuộc
                </label>
                <select
                  value={selectedParentId}
                  onChange={(e) => setSelectedParentId(e.target.value)}
                  className="w-full h-9 px-2 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900"
                >
                  <option value="">-- Phòng ban gốc (Level 1) --</option>
                  {allFlatDepts.filter(d => d.status === 'ACTIVE').map(d => (
                    <option key={d.id} value={d.id}>
                      {d.name} ({d.code})
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsCreateModalOpen(false)}
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
                  <span>Tạo phòng ban</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL UC-02: DI CHUYỂN PHÒNG BAN */}
      {isMoveModalOpen && selectedDept && (
        <div className="fixed inset-y-0 inset-x-0 z-50 flex items-center justify-center bg-slate-950/40 backdrop-blur-xs p-4">
          <div className="w-full max-w-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-xl overflow-hidden">
            <div className="flex items-center justify-between p-5 border-b border-slate-100 dark:border-slate-800">
              <div className="flex items-center gap-2">
                <ArrowRightLeft className="w-5 h-5 text-indigo-600" />
                <h3 className="font-semibold text-base text-slate-900 dark:text-slate-100">
                  Di Chuyển Phòng Ban (UC-02)
                </h3>
              </div>
              <button onClick={() => setIsMoveModalOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleExecuteMove} className="p-5 space-y-4">
              <div className="p-3 bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900 rounded-lg text-xs text-amber-800 dark:text-amber-300">
                <p className="font-semibold flex items-center gap-1.5 mb-1">
                  <AlertTriangle className="w-4 h-4 shrink-0" />
                  <span>Cập nhật đường dẫn Path & Di chuyển cây con</span>
                </p>
                Toàn bộ các phòng ban con và nhân sự trực thuộc sẽ được cập nhật lại chuỗi <strong>path</strong> tự động trong cùng một transaction nguyên tử.
              </div>

              <div>
                <span className="text-xs text-slate-500">Phòng ban cần di chuyển:</span>
                <p className="text-sm font-semibold text-slate-900 dark:text-slate-100 mt-0.5">
                  {selectedDept.name} ({selectedDept.code})
                </p>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                  Chọn phòng ban cha mới (ACTIVE)
                </label>
                <select
                  value={moveTargetParentId}
                  onChange={(e) => setMoveTargetParentId(e.target.value)}
                  className="w-full h-9 px-2 rounded-lg border border-slate-200 dark:border-slate-800 text-sm bg-white dark:bg-slate-900"
                >
                  <option value="">-- Chuyển thành phòng ban gốc (Level 1) --</option>
                  {allFlatDepts.filter(d => d.id !== selectedDept.id && d.status === 'ACTIVE').map(d => (
                    <option key={d.id} value={d.id}>
                      {d.name} ({d.code})
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsMoveModalOpen(false)}
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
                  <span>Xác nhận di chuyển</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
