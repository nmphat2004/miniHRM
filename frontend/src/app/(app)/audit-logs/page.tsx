"use client";

import React, { useState, useEffect } from "react";
import { History, ShieldCheck, RefreshCw, Search, Loader2 } from "lucide-react";
import { api, AuditLog } from "@/lib/api";

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
      setError(err.message || "Lỗi tải lịch sử kiểm toán");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const filteredLogs = logs.filter(l =>
    l.action.toLowerCase().includes(search.toLowerCase()) ||
    l.actorName.toLowerCase().includes(search.toLowerCase()) ||
    l.targetType.toLowerCase().includes(search.toLowerCase()) ||
    l.targetId.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2">
            <History className="w-6 h-6 text-indigo-600" />
            <h1 className="text-2xl font-bold font-heading text-slate-900 dark:text-slate-100">
              Sổ Cái Kiểm Toán (Audit Logs)
            </h1>
          </div>
          <p className="text-sm text-slate-500 mt-1">
            Ghi nhận nguyên tử toàn bộ biến động dữ liệu theo <strong>Ràng buộc 3</strong> (TransactWriteItems chỉ ghi thêm).
          </p>
        </div>

        <button onClick={loadData} className="p-2 rounded-lg border border-slate-200 dark:border-slate-800 text-slate-600 hover:bg-slate-50">
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
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
            className="w-full h-9 pl-9 pr-3 rounded-lg border text-sm bg-slate-50"
          />
        </div>
      </div>

      <div className="bg-white dark:bg-slate-900 border border-slate-200/80 rounded-2xl shadow-xs overflow-hidden">
        {isLoading ? (
          <div className="py-12 flex flex-col items-center justify-center text-slate-400 gap-2">
            <Loader2 className="w-6 h-6 animate-spin text-indigo-600" />
            <span className="text-sm">Đang tải sổ cái kiểm toán...</span>
          </div>
        ) : filteredLogs.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-sm">
            Không có bản ghi kiểm toán nào phù hợp.
          </div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-xs font-semibold uppercase text-slate-500 border-b">
              <tr>
                <th className="p-4 pl-6">Hành Động</th>
                <th className="p-4">Người Thực Hiện</th>
                <th className="p-4">Đối Tượng</th>
                <th className="p-4">Thời Gian Ghi Nhận</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filteredLogs.map(log => (
                <tr key={log.id} className="hover:bg-slate-50/50">
                  <td className="p-4 pl-6 font-mono font-semibold text-indigo-600">
                    {log.action}
                  </td>
                  <td className="p-4">
                    <div className="font-semibold text-slate-800">{log.actorName}</div>
                    <div className="font-mono text-xs text-slate-400">{log.actorId}</div>
                  </td>
                  <td className="p-4">
                    <span className="px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-700">
                      {log.targetType}
                    </span>
                    <div className="font-mono text-xs text-slate-400 mt-1">{log.targetId}</div>
                  </td>
                  <td className="p-4 font-mono text-xs text-slate-500">
                    {log.occurredAt}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
