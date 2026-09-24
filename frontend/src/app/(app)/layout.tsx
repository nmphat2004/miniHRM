"use client";

import React, { useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Network,
  Users,
  ShieldCheck,
  UserCircle,
  LogOut,
  Wifi,
  WifiOff,
  ChevronRight,
  Menu,
  X,
  Layers
} from "lucide-react";
import { api, clearAuth, getStoredUser } from "@/lib/api";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isOnline, setIsOnline] = useState(true);
  const [storedUser, setStoredUser] = useState<ReturnType<typeof getStoredUser>>(null);
  const [accountProfile, setAccountProfile] = useState<{ name?: string; avatar?: string }>({});

  useEffect(() => { setStoredUser(getStoredUser()); }, []);

  useEffect(() => {
    if (!storedUser?.userId) return;
    let cancelled = false;
    const refreshProfile = async () => {
      try {
        const employee = await api.getEmployeeById(storedUser.userId);
        if (!cancelled) setAccountProfile({ name: employee.fullName, avatar: employee.avatar });
      } catch {
        if (!cancelled) setAccountProfile({});
      }
    };
    const handleProfileUpdated = () => { void refreshProfile(); };
    void refreshProfile();
    window.addEventListener("minihrm:profile-updated", handleProfileUpdated);
    return () => {
      cancelled = true;
      window.removeEventListener("minihrm:profile-updated", handleProfileUpdated);
    };
  }, [storedUser?.userId, pathname]);

  useEffect(() => {
    const updateConnection = () => setIsOnline(window.navigator.onLine);
    updateConnection();
    window.addEventListener("online", updateConnection);
    window.addEventListener("offline", updateConnection);
    return () => {
      window.removeEventListener("online", updateConnection);
      window.removeEventListener("offline", updateConnection);
    };
  }, []);

  // Người dùng giả lập (Admin)
  const currentUser = {
    name: accountProfile.name || storedUser?.username || "Tài khoản Mini HRM",
    role: storedUser?.role === "admin" ? "Quản trị viên" : storedUser?.role === "manager" ? "Trưởng phòng" : "Nhân viên",
  };
  const nameParts = currentUser.name.trim().split(/\s+/);
  const initials = `${nameParts[0]?.[0] || "H"}${nameParts.length > 1 ? nameParts[nameParts.length - 1][0] : "R"}`.toUpperCase();

  const navItems = [
    {
      title: "Sơ đồ Phòng ban",
      href: "/departments",
      icon: Network,
      match: "/departments"
    },
    {
      title: "Danh sách Nhân sự",
      href: "/employees",
      icon: Users,
      match: "/employees"
    },
    {
      title: "Nhật ký Hệ thống",
      href: "/audit-logs",
      icon: ShieldCheck,
      match: "/audit-logs"
    },
  ];

  return (
    <div className="h-[100dvh] overflow-hidden flex flex-col md:flex-row bg-slate-50 dark:bg-slate-950">
      {/* Offline Alert Sticky Banner */}
      {!isOnline && (
        <div className="fixed top-0 left-0 right-0 z-50 bg-amber-500 text-white text-xs font-medium py-2 px-4 flex items-center justify-between shadow-md">
          <div className="flex items-center gap-2">
            <WifiOff className="w-4 h-4 animate-pulse" />
            <span>Mất kết nối tới máy chủ. Các thao tác ghi tạm thời bị vô hiệu hóa.</span>
          </div>
          <button onClick={() => window.location.reload()} className="underline hover:text-amber-100 text-xs font-semibold">Thử kết nối lại</button>
        </div>
      )}

      {/* Mobile Top Nav */}
      <div className="md:hidden flex items-center justify-between p-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-indigo-600 flex items-center justify-center text-white">
            <Layers className="w-4 h-4" />
          </div>
          <span className="font-bold font-heading text-slate-900 dark:text-slate-100">Mini HRM</span>
        </div>
        <button
          onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
          className="p-2 text-slate-600 dark:text-slate-300"
        >
          {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
        </button>
      </div>

      {/* Sidebar Navigation */}
      {isMobileMenuOpen && <button aria-label="Đóng menu" onClick={() => setIsMobileMenuOpen(false)} className="fixed inset-0 z-30 bg-slate-950/35 md:hidden" />}
      <aside
        className={`fixed inset-y-0 left-0 z-40 w-64 bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 flex flex-col transition-transform duration-200 md:translate-x-0 md:static md:shrink-0 ${
          isMobileMenuOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        {/* Brand */}
        <div className="h-16 flex items-center gap-3 px-6 border-b border-slate-100 dark:border-slate-800/80">
          <div className="w-9 h-9 rounded-xl bg-indigo-600 flex items-center justify-center text-white shadow-sm">
            <Layers className="w-5 h-5" />
          </div>
          <div>
            <span className="font-bold font-heading text-slate-900 dark:text-slate-100 tracking-tight text-base block">
              Mini HRM
            </span>
            <span className="text-[10px] uppercase font-mono tracking-widest text-indigo-600 dark:text-indigo-400 font-semibold block -mt-0.5">
              Enterprise v1.0
            </span>
          </div>
        </div>

        {/* Nav Items */}
        <div className="flex-1 py-6 px-3 space-y-1 overflow-y-auto">
          <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider px-3 mb-2">
            Quản trị tổ chức
          </div>
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = pathname.startsWith(item.match);
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setIsMobileMenuOpen(false)}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all ${
                  isActive
                    ? "relative bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-300 font-semibold shadow-xs before:absolute before:inset-y-2 before:left-0 before:w-0.5 before:rounded-full before:bg-indigo-600"
                    : "text-slate-600 dark:text-slate-400 hover:translate-x-0.5 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900"
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? "text-indigo-600 dark:text-indigo-400" : "text-slate-400"}`} />
                <span>{item.title}</span>
              </Link>
            );
          })}
        </div>

        {/* User Card & Logout */}
        <div className="p-4 border-t border-slate-100 dark:border-slate-800">
          <div className="flex items-center gap-3 mb-3 p-2 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-100 dark:border-slate-800">
            {accountProfile.avatar ? (
              <Image src={accountProfile.avatar} alt={`Ảnh đại diện ${currentUser.name}`} width={36} height={36} unoptimized className="h-9 w-9 shrink-0 rounded-lg object-cover" />
            ) : (
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-100 text-sm font-bold text-indigo-700">{initials}</div>
            )}
            <div className="min-w-0 flex-1">
              <p className="text-xs font-semibold text-slate-900 dark:text-slate-100 truncate">
                {currentUser.name}
              </p>
              <div className="flex items-center gap-1.5 mt-0.5">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                <span className="text-[10px] text-slate-500 dark:text-slate-400 truncate">
                  {currentUser.role}
                </span>
              </div>
            </div>
          </div>

          <button
            onClick={async () => {
              try { await api.logout(); } finally { clearAuth(); router.push("/login"); }
            }}
            className="w-full flex items-center justify-center gap-2 py-2 rounded-md text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-all"
          >
            <LogOut className="w-3.5 h-3.5" />
            <span>Đăng xuất phiên làm việc</span>
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-w-0 min-h-0 overflow-y-auto">
        {/* Top bar header */}
        <header className="sticky top-0 z-20 h-16 hidden md:flex shrink-0 items-center justify-between px-8 bg-white/95 dark:bg-slate-900/95 backdrop-blur border-b border-slate-200 dark:border-slate-800">
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <span>Tổ chức</span>
            <ChevronRight className="w-3.5 h-3.5 text-slate-400" />
            <span className="text-slate-800 dark:text-slate-200 font-medium">
              Công ty CP Công Nghệ Mini HRM
            </span>
          </div>
          <div className="flex items-center gap-4 text-xs">
            <div className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-medium bg-emerald-50 dark:bg-emerald-950/40 px-2.5 py-1 rounded-full border border-emerald-200 dark:border-emerald-900">
              <Wifi className="w-3 h-3" />
              <span>Máy chủ hoạt động (Local Go API)</span>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <div className="p-4 md:p-8 flex-1 min-h-full">
          {children}
        </div>
      </main>
    </div>
  );
}
