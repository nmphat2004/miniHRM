"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AlertCircle, ArrowUpRight, RefreshCw, Search, Users, WifiOff } from "lucide-react";
import { api, ApiError, apiErrorMessage, Department, Employee } from "@/lib/api";
import { useOnline } from "@/lib/use-online";

type StatusFilter = "ACTIVE" | "RESIGNED" | "ALL";
const pageSize = 8;

export default function DepartmentEmployees({ department, departments }: { department: Department | null; departments: Department[] }) {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [status, setStatus] = useState<StatusFilter>("ALL");
  const [includeDescendants, setIncludeDescendants] = useState(true);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [refresh, setRefresh] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [noAccess, setNoAccess] = useState(false);
  const online = useOnline();

  useEffect(() => {
    setPage(1);
    setSearch("");
    setStatus("ALL");
  }, [department?.id]);

  useEffect(() => {
    if (!department) return;
    let cancelled = false;
    setLoading(true);
    setError(null);
    setNoAccess(false);
    api.getEmployees(department.id, status, undefined, includeDescendants)
      .then((data) => { if (!cancelled) setEmployees(data || []); })
      .catch((err: unknown) => {
        if (cancelled) return;
        setEmployees([]);
        setNoAccess(err instanceof ApiError && (err.status === 403 || err.status === 404));
        setError(apiErrorMessage(err, "Không tải được nhân viên của phòng ban."));
      })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [department?.id, status, includeDescendants, refresh]);

  if (!department) {
    return (
      <section className="flex min-h-72 flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-white px-6 text-center dark:border-slate-700 dark:bg-slate-900">
        <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-950/60"><Users className="h-5 w-5" /></div>
        <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Chọn một phòng ban</h2>
        <p className="mt-1 max-w-xs text-xs leading-5 text-slate-500">Nhấn tên phòng ban để xem nhân viên trực thuộc. Mũi tên chỉ dùng để mở hoặc thu gọn nhánh cây.</p>
      </section>
    );
  }

  const filtered = employees.filter((employee) => {
    const term = search.trim().toLocaleLowerCase("vi-VN");
    return !term || [employee.fullName, employee.code, employee.title].some((value) => value.toLocaleLowerCase("vi-VN").includes(term));
  }).sort((a, b) => a.fullName.localeCompare(b.fullName, "vi-VN") || a.id.localeCompare(b.id));
  const pageCount = Math.max(1, Math.ceil(filtered.length / pageSize));
  const visible = filtered.slice((page - 1) * pageSize, page * pageSize);

  return (
    <section aria-label={`Nhân viên thuộc ${department.name}`} className="min-w-0 overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm shadow-slate-200/40 dark:border-slate-800 dark:bg-slate-900 dark:shadow-none">
      <div className="border-b border-slate-100 p-4 dark:border-slate-800 sm:p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="text-xs font-medium text-indigo-600 dark:text-indigo-400">Nhân sự theo phòng ban</p>
            <h2 className="mt-1 text-lg font-semibold tracking-tight text-slate-900 dark:text-slate-100">{department.name}</h2>
            <p className="mt-1 text-xs text-slate-500">{department.employeeCount} nhân viên trong nhánh · {department.directEmployeeCount ?? 0} trực tiếp</p>
            <p className="mt-1 text-xs text-slate-500">Trưởng phòng: {department.managerName || "Chưa bổ nhiệm"}</p>
          </div>
          <button type="button" onClick={() => setRefresh((value) => value + 1)} aria-label="Tải lại nhân viên" title="Tải lại nhân viên" className="rounded-lg border border-slate-200 p-2 text-slate-600 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"><RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} /></button>
        </div>
        <div className="mt-4 inline-flex rounded-lg border border-slate-200 p-1 dark:border-slate-700" role="group" aria-label="Phạm vi nhân viên">
          <button type="button" aria-pressed={includeDescendants} onClick={() => { setIncludeDescendants(true); setPage(1); }} className={`rounded-md px-3 py-1.5 text-xs font-medium ${includeDescendants ? "bg-indigo-600 text-white" : "text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"}`}>Cả nhánh</button>
          <button type="button" aria-pressed={!includeDescendants} onClick={() => { setIncludeDescendants(false); setPage(1); }} className={`rounded-md px-3 py-1.5 text-xs font-medium ${!includeDescendants ? "bg-indigo-600 text-white" : "text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"}`}>Chỉ phòng này</button>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          {(["ACTIVE", "RESIGNED", "ALL"] as const).map((option) => (
            <button key={option} type="button" aria-pressed={status === option} onClick={() => { setStatus(option); setPage(1); }} className={`whitespace-nowrap rounded-md border px-3 py-1.5 text-xs font-medium ${status === option ? "border-indigo-200 bg-indigo-50 text-indigo-700 dark:border-indigo-800 dark:bg-indigo-950 dark:text-indigo-300" : "border-slate-200 text-slate-600 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"}`}>
              {option === "ACTIVE" ? "Đang làm" : option === "RESIGNED" ? "Đã nghỉ" : "Tất cả"}
            </button>
          ))}
        </div>
        <label className="relative mt-3 block">
          <span className="sr-only">Tìm nhân viên trong phòng ban</span>
          <Search className="pointer-events-none absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
          <input value={search} onChange={(event) => { setSearch(event.target.value); setPage(1); }} placeholder="Tìm tên, mã hoặc chức danh..." className="h-9 w-full rounded-lg border border-slate-200 bg-slate-50 pl-9 pr-3 text-sm dark:border-slate-700 dark:bg-slate-950" />
        </label>
      </div>

      {!online && <p className="flex items-center gap-2 border-b border-amber-200 bg-amber-50 px-4 py-2 text-xs text-amber-800"><WifiOff className="h-4 w-4 shrink-0" />Mất kết nối. Danh sách có thể chưa được cập nhật.</p>}
      {loading ? (
        <div aria-label="Đang tải nhân viên" className="space-y-3 p-4">{[0, 1, 2].map((index) => <div key={index} className="h-14 animate-pulse rounded-lg bg-slate-100 dark:bg-slate-800" />)}</div>
      ) : error ? (
        <div role="alert" className="p-6 text-sm text-slate-600 dark:text-slate-300">
          <AlertCircle className="mb-2 h-5 w-5 text-rose-600" />
          <p>{noAccess ? "Bạn không có quyền xem danh sách nhân viên của phòng ban này." : error}</p>
          {!noAccess && <button type="button" onClick={() => setRefresh((value) => value + 1)} className="mt-3 text-sm font-medium text-indigo-600 hover:underline">Thử lại</button>}
        </div>
      ) : filtered.length === 0 ? (
        <div className="px-6 py-10 text-center"><Users className="mx-auto h-7 w-7 text-slate-300" /><p className="mt-2 text-sm font-medium text-slate-700 dark:text-slate-200">{search ? "Không tìm thấy nhân viên phù hợp" : "Chưa có nhân viên ở trạng thái này"}</p><p className="mt-1 text-xs text-slate-500">{search ? "Thử tìm bằng tên hoặc mã khác." : "Chọn trạng thái khác để xem thêm."}</p></div>
      ) : (
        <>
          <div className="divide-y divide-slate-100 dark:divide-slate-800">
            {visible.map((employee) => (
              <div key={employee.id} className="flex min-w-0 items-center justify-between gap-3 px-4 py-3 hover:bg-slate-50/70 dark:hover:bg-slate-800/40 sm:px-5">
                <div className="min-w-0"><p className="truncate text-sm font-semibold text-slate-900 dark:text-slate-100">{employee.fullName}{employee.status === "RESIGNED" && <span className="ml-2 text-xs font-normal text-slate-500">Đã nghỉ</span>}</p><p className="mt-0.5 truncate text-xs text-slate-500">{employee.code} · {employee.title || "Chưa có chức danh"}</p>{includeDescendants && employee.departmentId !== department.id && <p className="mt-0.5 truncate text-xs text-indigo-600 dark:text-indigo-400">{departments.find((item) => item.id === employee.departmentId)?.name || employee.departmentName}</p>}</div>
                <Link href={`/employees/${employee.id}`} aria-label={`Xem hồ sơ ${employee.fullName}`} className="inline-flex shrink-0 items-center gap-1 whitespace-nowrap rounded-lg px-2 py-1.5 text-xs font-medium text-indigo-600 hover:bg-indigo-50 dark:text-indigo-300 dark:hover:bg-indigo-950">Hồ sơ <ArrowUpRight className="h-3.5 w-3.5" /></Link>
              </div>
            ))}
          </div>
          <div className="flex flex-wrap items-center justify-between gap-2 border-t border-slate-100 px-4 py-3 text-xs text-slate-500 dark:border-slate-800 sm:px-5">
            <span>Hiển thị {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, filtered.length)} / {filtered.length} nhân viên</span>
            {pageCount > 1 && <div className="flex items-center gap-2"><button type="button" disabled={page === 1} onClick={() => setPage((value) => value - 1)} className="rounded-md border px-2 py-1 disabled:opacity-40">Trước</button><span>{page}/{pageCount}</span><button type="button" disabled={page === pageCount} onClick={() => setPage((value) => value + 1)} className="rounded-md border px-2 py-1 disabled:opacity-40">Sau</button></div>}
          </div>
        </>
      )}
    </section>
  );
}
