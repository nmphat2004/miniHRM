"use client";

import React, { useState, useEffect, useRef } from "react";
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
  Columns3,
  AlertTriangle,
  CheckCircle2,
  X,
  Loader2,
  UserCheck,
  RefreshCw
} from "lucide-react";
import { api, apiErrorMessage, Department, getStoredUser } from "@/lib/api";
import { useOnline } from "@/lib/use-online";
import DepartmentEmployees from "./department-employees";

export default function DepartmentsPage() {
  const [departments, setDepartments] = useState<Department[]>([]);
  const [allFlatDepts, setAllFlatDepts] = useState<Department[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Expanded nodes state
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [treeView, setTreeView] = useState<"tree" | "columns">("tree");
  const [selectedDeptId, setSelectedDeptId] = useState<string | null>(null);
  const employeesPanelRef = useRef<HTMLDivElement>(null);

  // Modals state
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isMoveModalOpen, setIsMoveModalOpen] = useState(false);
  const [selectedDept, setSelectedDept] = useState<Department | null>(null);
  const [deptAction, setDeptAction] = useState<"rename" | "manager" | null>(null);
  const [actionDept, setActionDept] = useState<Department | null>(null);
  const [actionValue, setActionValue] = useState("");
  const [managerChoices, setManagerChoices] = useState<{ id: string; fullName: string; departmentPath?: string }[]>([]);

  // Form states
  const [newCode, setNewCode] = useState("");
  const [newName, setNewName] = useState("");
  const [selectedParentId, setSelectedParentId] = useState("");
  const [moveTargetParentId, setMoveTargetParentId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const currentUser = getStoredUser();
  const isOnline = useOnline();

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
      setSelectedDeptId((id) => id && (flat || []).some((dept) => dept.id === id) ? id : null);

      // Auto expand root nodes
      const expMap: Record<string, boolean> = {};
      (flat || []).forEach(d => { expMap[d.id] = true; });
      setExpanded(expMap);
    } catch (err: any) {
      setError(apiErrorMessage(err, "Không thể tải cây phòng ban"));
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

  const selectDepartment = (id: string) => {
    setSelectedDeptId(id);
    if (window.innerWidth < 1024) {
      requestAnimationFrame(() => employeesPanelRef.current?.scrollIntoView({ behavior: "smooth", block: "start" }));
    }
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
      setError(apiErrorMessage(err, "Lỗi khi tạo phòng ban"));
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
      setError(apiErrorMessage(err, "Lỗi di chuyển phòng ban"));
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

  const openDeptAction = async (dept: Department, action: "rename" | "manager") => {
    setActionDept(dept);
    setDeptAction(action);
    setActionValue(action === "rename" ? dept.name : "");
    if (action === "manager") {
      try {
        const employees = await api.getEmployees(undefined, "ACTIVE");
        setManagerChoices(employees.map(({ id, fullName, departmentPath }) => ({ id, fullName, departmentPath })));
      } catch (err: any) {
        setError(err.message || "Không tải được danh sách nhân viên");
      }
    }
  };

  const handleDeptAction = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!actionDept || !deptAction) return;
    setIsSubmitting(true);
    try {
      if (deptAction === "rename") await api.updateDepartment(actionDept.id, actionValue, actionDept.version);
      else await api.assignDepartmentManager(actionDept.id, actionValue);
      setDeptAction(null);
      setActionDept(null);
      setSuccessMsg(deptAction === "rename" ? "Đã cập nhật tên phòng ban." : "Đã bổ nhiệm trưởng phòng.");
      await loadData();
    } catch (err: any) {
      setError(apiErrorMessage(err, "Không thể lưu thay đổi phòng ban"));
    } finally {
      setIsSubmitting(false);
    }
  };

  const renderTree = (nodes: Department[], depth: number = 1) => {
    return nodes.map((dept) => {
      const isExp = expanded[dept.id] ?? true;
      const hasChildren = dept.children && dept.children.length > 0;
      const isArchived = dept.status === "ARCHIVED";

      return (
          <div key={dept.id} className="relative space-y-1">
          <div
              className={`group flex items-center justify-between gap-2 p-3 rounded-xl border transition-all ${
              selectedDeptId === dept.id
                ? "border-indigo-400 bg-indigo-50/70 ring-1 ring-indigo-200 dark:border-indigo-600 dark:bg-indigo-950/50"
                : isArchived
                  ? "bg-slate-50/60 dark:bg-slate-900/40 border-slate-200/60 dark:border-slate-800/60 opacity-70"
                  : "bg-white dark:bg-slate-900 border-slate-200/80 dark:border-slate-800 hover:border-indigo-300 dark:hover:border-indigo-800 shadow-xs"
            }`}
            style={{ marginLeft: `${(depth - 1) * 28}px` }}
          >
            <div className="flex min-w-0 flex-1 items-center gap-2.5">
              <button
                type="button"
                onClick={() => toggleExpand(dept.id)}
                disabled={!hasChildren}
                aria-label={hasChildren ? `${isExp ? "Thu gọn" : "Mở rộng"} ${dept.name}` : undefined}
                aria-expanded={hasChildren ? isExp : undefined}
                className="w-6 h-7 shrink-0 flex items-center justify-center text-slate-400 hover:text-slate-600 rounded disabled:cursor-default"
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

              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <button type="button" aria-pressed={selectedDeptId === dept.id} onClick={() => selectDepartment(dept.id)} className="min-w-0 truncate text-left text-sm font-semibold text-slate-900 hover:text-indigo-700 dark:text-slate-100 dark:hover:text-indigo-300">
                    {dept.name}
                  </button>
                  <span className={`px-2 py-0.5 rounded-full border text-[10px] font-medium shrink-0 ${[
                    "bg-indigo-50 text-indigo-700 border-indigo-200",
                    "bg-blue-50 text-blue-700 border-blue-200",
                    "bg-teal-50 text-teal-700 border-teal-200",
                    "bg-violet-50 text-violet-700 border-violet-200",
                    "bg-purple-50 text-purple-700 border-purple-200",
                  ][Math.min(depth, 5) - 1]}`}>Cấp {depth}/5</span>
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
              <span title={`${dept.employeeCount} nhân viên cả nhánh (gồm ${dept.directEmployeeCount ?? 0} trực tiếp, kể cả đã nghỉ)`} className="flex items-center gap-1 text-xs text-slate-500 px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800">
                <Users className="w-3.5 h-3.5" />
                <strong>{dept.employeeCount}</strong> nhân sự
              </span>

              {(currentUser?.role === 'admin' || currentUser?.role === 'manager') && (
                <div className="flex items-center gap-1 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 lg:group-focus-within:opacity-100 transition-opacity">
                  <button onClick={() => openDeptAction(dept, "rename")} title="Đổi tên phòng ban" className="p-1.5 rounded-lg text-slate-500 hover:text-indigo-600 hover:bg-indigo-50"><Edit2 className="w-4 h-4" /></button>
                  {currentUser?.role === 'admin' && <button onClick={() => openDeptAction(dept, "manager")} title="Bổ nhiệm trưởng phòng" className="p-1.5 rounded-lg text-slate-500 hover:text-indigo-600 hover:bg-indigo-50"><UserCheck className="w-4 h-4" /></button>}
                  {currentUser?.role === 'admin' && <>
                  {!isArchived && (
                    <button
                      onClick={() => {
                        setSelectedDept(dept);
                        setIsMoveModalOpen(true);
                      }}
                      disabled={!isOnline}
                      title="Di chuyển phòng ban (UC-02)"
                      className="p-1.5 rounded-lg text-slate-500 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-950/50"
                    >
                      <ArrowRightLeft className="w-4 h-4" />
                    </button>
                  )}
                  {!isArchived && (
                    <button
                      onClick={() => handleArchive(dept)}
                      disabled={!isOnline}
                      title="Lưu trữ phòng ban (BR-PB-06)"
                      className="p-1.5 rounded-lg text-slate-500 hover:text-amber-600 hover:bg-amber-50 dark:hover:bg-amber-950/50"
                    >
                      <Archive className="w-4 h-4" />
                    </button>
                  )}
                  </>}
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

  const departmentLevels: Department[][] = [];
  const collectLevels = (nodes: Department[], depth = 0) => {
    if (!nodes.length) return;
    departmentLevels[depth] = [...(departmentLevels[depth] || []), ...nodes];
    nodes.forEach((node) => collectLevels(node.children || [], depth + 1));
  };
  collectLevels(departments);

  const parentNames = new Map(allFlatDepts.map((dept) => [dept.id, dept.name]));
  const selectedDepartment = allFlatDepts.find((dept) => dept.id === selectedDeptId) || null;

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Network className="w-6 h-6 text-indigo-600" />
            <h1 className="text-2xl font-bold font-heading tracking-tight text-slate-900 dark:text-slate-100">
              Sơ đồ phòng ban
            </h1>
          </div>
          <p className="text-sm text-slate-500 mt-1">
            Quản trị cấu trúc tổ chức phân cấp, di chuyển phòng ban (UC-02) và phân quyền Scoping (Ràng buộc 1).
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-1">
            <button
              type="button"
              onClick={() => setTreeView("tree")}
              aria-pressed={treeView === "tree"}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${treeView === "tree" ? "bg-indigo-600 text-white" : "text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800"}`}
            >
              <FolderTree className="w-3.5 h-3.5" /> Dạng cây
            </button>
            <button
              type="button"
              onClick={() => setTreeView("columns")}
              aria-pressed={treeView === "columns"}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${treeView === "columns" ? "bg-indigo-600 text-white" : "text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800"}`}
            >
              <Columns3 className="w-3.5 h-3.5" /> Chia cột
            </button>
          </div>
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
              disabled={!isOnline}
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

      {/* Sơ đồ và danh sách nhân viên của đơn vị được chọn */}
      <div className="grid min-w-0 items-start gap-5 lg:grid-cols-[minmax(0,1.1fr)_minmax(360px,0.9fr)]">
      <section aria-label="Sơ đồ phòng ban" className="min-w-0 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-sm shadow-slate-200/40 dark:border-slate-800 dark:bg-slate-900 dark:shadow-none sm:p-5">
        <p className="mb-3 text-xs text-slate-500">Số nhân sự tính cả nhánh và người đã nghỉ · Chọn tên phòng để xem danh sách</p>
        {isLoading ? (
          <div aria-label="Đang tải sơ đồ phòng ban" className="space-y-3 py-2">
            {["w-full", "w-[94%]", "w-[88%]", "w-[82%]", "w-[76%]"].map((width, index) => <div key={index} className={`h-[60px] ${width} animate-pulse rounded-xl border border-slate-100 bg-slate-100 dark:border-slate-800 dark:bg-slate-800`} />)}
          </div>
        ) : departments.length === 0 ? (
          <div className="flex flex-col items-center py-14 text-center">
            <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-indigo-50 text-indigo-600 dark:bg-indigo-950/60 dark:text-indigo-300"><FolderTree className="h-5 w-5" /></div>
            <p className="text-sm font-semibold text-slate-800 dark:text-slate-200">Chưa có phòng ban trong phạm vi này</p>
            <p className="mt-1 max-w-sm text-xs leading-5 text-slate-500">Khi có đơn vị được tạo hoặc phân quyền cho bạn, sơ đồ sẽ xuất hiện tại đây.</p>
          </div>
        ) : (
          treeView === "tree" ? (
            <div className="overflow-x-auto pb-2"><div className="min-w-[500px] space-y-2">{renderTree(departments)}</div></div>
          ) : (
            <div className="space-y-5">
                {departmentLevels.map((level, index) => (
                  <section key={index} className="min-w-0">
                    <div className="mb-3 rounded-lg bg-slate-100 px-3 py-2 text-xs font-semibold text-slate-600 dark:bg-slate-800 dark:text-slate-300">
                      Cấp {index + 1} <span className="ml-1 text-slate-400">({level.length})</span>
                    </div>
                    <div className="grid grid-cols-1 items-start gap-3 sm:grid-cols-2">
                    {level.map((dept) => (
                      <article key={dept.id} className={`min-w-0 rounded-xl border p-3 transition-colors ${selectedDeptId === dept.id ? "border-indigo-400 bg-indigo-50/70 ring-1 ring-indigo-200 dark:bg-indigo-950/50" : dept.status === "ARCHIVED" ? "border-slate-200 bg-slate-50 dark:bg-slate-900" : "border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-900"}`}>
                        <div className="flex items-start justify-between gap-2">
                          <div className="min-w-0">
                            <h3><button type="button" aria-pressed={selectedDeptId === dept.id} onClick={() => selectDepartment(dept.id)} className="max-w-full truncate text-left text-sm font-semibold text-slate-900 hover:text-indigo-700 dark:text-slate-100 dark:hover:text-indigo-300">{dept.name}</button></h3>
                            <p className="mt-1 font-mono text-[11px] text-slate-500">{dept.code}</p>
                          </div>
                          <span title={`${dept.employeeCount} nhân viên cả nhánh (gồm ${dept.directEmployeeCount ?? 0} trực tiếp, kể cả đã nghỉ)`} className="shrink-0 whitespace-nowrap rounded-full bg-indigo-50 px-2 py-0.5 text-[10px] font-medium text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300">{dept.employeeCount} nhân sự</span>
                        </div>
                        <div className="mt-3 space-y-1 border-t border-slate-100 pt-2 text-xs dark:border-slate-800">
                          <p className="truncate text-slate-500">{dept.parentId ? `Thuộc: ${parentNames.get(dept.parentId) || "Phòng cha"}` : "Đơn vị gốc"}</p>
                          <p className="truncate text-indigo-600 dark:text-indigo-400">{dept.managerName || "Chưa có Trưởng phòng"}</p>
                          {dept.status === "ARCHIVED" && <p className="text-amber-600">Đã lưu trữ</p>}
                        </div>
                      </article>
                    ))}
                    </div>
                  </section>
                ))}
            </div>
          )
        )}
      </section>
      <div ref={employeesPanelRef} className="min-w-0 scroll-mt-4 lg:sticky lg:top-20"><DepartmentEmployees department={selectedDepartment} departments={allFlatDepts} /></div>
      </div>

      {deptAction && actionDept && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4">
          <form onSubmit={handleDeptAction} className="w-full max-w-md space-y-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-xl dark:border-slate-700 dark:bg-slate-900">
            <div className="flex items-center justify-between">
              <h2 className="font-semibold">{deptAction === "rename" ? "Đổi tên phòng ban" : "Bổ nhiệm trưởng phòng"}</h2>
              <button type="button" onClick={() => setDeptAction(null)} aria-label="Đóng"><X className="h-5 w-5" /></button>
            </div>
            <p className="text-sm text-slate-500">{actionDept.name} ({actionDept.code})</p>
            {deptAction === "rename" ? (
              <input autoFocus required minLength={2} maxLength={100} value={actionValue} onChange={(event) => setActionValue(event.target.value)} className="h-10 w-full rounded-lg border px-3 text-sm dark:bg-slate-950" />
            ) : (
              <select required value={actionValue} onChange={(event) => setActionValue(event.target.value)} className="h-10 w-full rounded-lg border px-3 text-sm dark:bg-slate-950">
                <option value="">-- Chọn nhân viên ACTIVE --</option>
                {managerChoices.filter((employee) => employee.departmentPath?.startsWith(actionDept.path)).map((employee) => <option key={employee.id} value={employee.id}>{employee.fullName}</option>)}
              </select>
            )}
            <div className="flex justify-end gap-2">
              <button type="button" onClick={() => setDeptAction(null)} className="rounded-lg border px-3 py-2 text-sm">Hủy</button>
              <button disabled={isSubmitting || !isOnline} className="rounded-lg bg-indigo-600 px-3 py-2 text-sm text-white disabled:opacity-50">{isSubmitting ? "Đang lưu…" : "Lưu"}</button>
            </div>
          </form>
        </div>
      )}

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
                  disabled={isSubmitting || !isOnline}
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
                  disabled={isSubmitting || !isOnline}
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
