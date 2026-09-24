"use client";

import React, { useState, useEffect } from "react";
import {
  History,
  RefreshCw,
  Search,
  UserRound,
  ArrowRight,
} from "lucide-react";
import { api, apiErrorMessage, AuditLog, getAuditActionLabel } from "@/lib/api";

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [search, setSearch] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await api.getAuditLogs();
      setLogs(data || []);
    } catch (err: any) {
      setError(apiErrorMessage(err, "Lỗi tải lịch sử kiểm toán"));
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const filteredLogs = logs.filter(
    (l) =>
      l.action.toLowerCase().includes(search.toLowerCase()) ||
      l.actorName.toLowerCase().includes(search.toLowerCase()) ||
      l.targetType.toLowerCase().includes(search.toLowerCase()) ||
      l.targetId.toLowerCase().includes(search.toLowerCase()),
  );

  const formatValue = (value: unknown) => {
    if (value === null || value === undefined || value === "") return "—";
    if (typeof value === "boolean") return value ? "Có" : "Không";
    if (typeof value === "object") return JSON.stringify(value);
    return String(value);
  };
  const fieldLabels: Record<string, string> = {
    name: "Tên phòng ban", fullName: "Họ tên", title: "Chức danh", email: "Email",
    departmentName: "Phòng ban", status: "Trạng thái", joinedAt: "Ngày vào làm",
    resignedAt: "Ngày nghỉ việc", resignReason: "Lý do nghỉ việc", managerName: "Trưởng phòng",
    code: "Mã nghiệp vụ", avatar: "Ảnh đại diện",
  };
  const visibleDiffFields = (log: AuditLog) =>
    Array.from(new Set([...Object.keys(log.before || {}), ...Object.keys(log.after || {})]))
      .filter((field) => !/(^|[A-Z])id$/i.test(field) && field !== "path" && field !== "version");

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <History className="w-6 h-6 text-indigo-600" />
            <h1 className="text-2xl font-bold font-heading tracking-tight text-slate-900 dark:text-slate-100">
              Nhật ký hệ thống
            </h1>
          </div>
          <p className="text-sm text-slate-500 mt-1">
            Ghi nhận nguyên tử toàn bộ biến động dữ liệu theo{" "}
            <strong>Ràng buộc 3</strong> (TransactWriteItems chỉ ghi thêm).
          </p>
        </div>

        <button
          onClick={loadData}
          className="p-2 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-600 hover:bg-slate-50"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? "animate-spin" : ""}`} />
        </button>
      </div>

      <div className="p-4 bg-white dark:bg-slate-900 border border-slate-200/80 rounded-xl">
        <div className="relative max-w-md">
          <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-400" />
          <input
            type="text"
            placeholder="Tìm theo hành động, người thực hiện, đối tượng..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-9 pl-9 pr-3 rounded-lg border border-slate-200 bg-slate-50 text-sm dark:border-slate-700 dark:bg-slate-950"
          />
        </div>
      </div>

      <div className="bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 rounded-2xl shadow-sm shadow-slate-200/40 dark:shadow-none overflow-hidden">
        {isLoading ? (
          <div
            aria-label="Đang tải nhật ký"
            className="divide-y divide-slate-100 p-5 dark:divide-slate-800"
          >
            {[0, 1, 2, 3].map((row) => (
              <div
                key={row}
                className="grid grid-cols-1 gap-3 py-5 sm:grid-cols-3 sm:gap-5"
              >
                <div className="h-8 w-36 animate-pulse rounded bg-slate-100 dark:bg-slate-800" />
                <div className="h-8 w-40 animate-pulse rounded bg-slate-100 dark:bg-slate-800" />
                <div className="h-8 w-52 animate-pulse rounded bg-slate-100 dark:bg-slate-800" />
              </div>
            ))}
          </div>
        ) : filteredLogs.length === 0 ? (
          <div className="flex flex-col items-center py-14 text-center">
            <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-indigo-50 text-indigo-600 dark:bg-indigo-950/60 dark:text-indigo-300">
              <History className="h-5 w-5" />
            </div>
            <p className="text-sm font-semibold text-slate-800 dark:text-slate-200">
              Chưa có nhật ký phù hợp
            </p>
            <p className="mt-1 text-xs text-slate-500">
              Thử tìm theo người thực hiện, hành động hoặc đối tượng.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px] text-left text-sm">
              <thead className="bg-slate-50 text-xs font-semibold uppercase text-slate-500 border-b">
                <tr>
                  <th className="p-4 pl-6">Người thực hiện</th>
                  <th className="p-4">Hành động</th>
                  <th className="p-4">Đối tượng bị tác động</th>
                  <th className="p-4">Thời gian</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {filteredLogs.map((log) => (
                  <tr key={log.id} className="hover:bg-slate-50/50">
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <span className="w-8 h-8 rounded-full bg-indigo-50 text-indigo-600 flex items-center justify-center">
                          <UserRound className="w-4 h-4" />
                        </span>
                        <div>
                          <div className="font-semibold text-slate-800">
                            {log.actorName || "Không rõ người thực hiện"}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="p-4">
                      <div className="font-semibold text-indigo-700">{getAuditActionLabel(log.action)}</div>
                    </td>
                    <td className="p-4">
                      <div className="font-medium text-slate-800">
                        {log.targetType === "Department"
                          ? "Phòng ban"
                          : log.targetType === "Employee"
                            ? "Nhân viên"
                            : log.targetType}
                      </div>
                      <div className="mt-1 text-xs text-slate-600">
                        {log.targetName || log.after?.fullName || log.before?.fullName || log.after?.name || log.before?.name || "Thông tin đối tượng"}
                      </div>
                      {visibleDiffFields(log).length > 0 ? (
                        <div className="mt-2 space-y-1.5">
                          {visibleDiffFields(log).map((field) => (
                            <div
                              key={field}
                              className="text-xs flex flex-wrap items-center gap-1.5"
                            >
                              <span className="text-slate-500">{fieldLabels[field] || "Thông tin"}:</span>
                              <span className="rounded bg-rose-50 text-rose-700 px-1.5 py-0.5 line-through">
                                {formatValue(log.before?.[field])}
                              </span>
                              <ArrowRight className="w-3 h-3 text-slate-400" />
                              <span className="rounded bg-emerald-50 text-emerald-700 px-1.5 py-0.5">
                                {formatValue(log.after?.[field])}
                              </span>
                            </div>
                          ))}
                        </div>
                      ) : null}
                    </td>
                    <td className="p-4 text-xs text-slate-500">
                      {new Date(log.occurredAt).toLocaleString("vi-VN", { dateStyle: "medium", timeStyle: "short" })}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
