// API Client kết nối trực tiếp với Mini HRM Go Backend tại http://localhost:8080/api/v1
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface UserClaims {
  userId: string;
  username: string;
  role: 'admin' | 'manager' | 'employee';
  departmentId?: string;
  departmentPath?: string;
}

export interface Department {
  id: string;
  code: string;
  name: string;
  parentId: string;
  path: string;
  managerId?: string;
  managerName?: string;
  status: 'ACTIVE' | 'ARCHIVED' | 'MOVING';
  version: number;
  employeeCount: number;
  createdAt: string;
  updatedAt: string;
  children?: Department[];
}

export interface Employee {
  id: string;
  code: string;
  fullName: string;
  email: string;
  departmentId: string;
  departmentName?: string;
  departmentPath?: string;
  title: string;
  joinedAt: string;
  status: 'ACTIVE' | 'RESIGNED';
  version: number;
  resignedAt?: string;
  resignReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuditLog {
  id: string;
  actorId: string;
  actorName: string;
  action: string;
  targetType: string;
  targetId: string;
  before?: Record<string, any>;
  after?: Record<string, any>;
  occurredAt: string;
}

// Lấy auth token lưu trong localStorage (hoặc cookie)
export function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('minihrm_token');
}

export function setAuthToken(token: string, user: UserClaims) {
  if (typeof window === 'undefined') return;
  localStorage.setItem('minihrm_token', token);
  localStorage.setItem('minihrm_user', JSON.stringify(user));
}

export function getStoredUser(): UserClaims | null {
  if (typeof window === 'undefined') return null;
  const raw = localStorage.getItem('minihrm_user');
  if (!raw) return null;
  try { return JSON.parse(raw); } catch { return null; }
}

export function clearAuth() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('minihrm_token');
  localStorage.removeItem('minihrm_user');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  });

  const body = await res.json();
  if (!res.ok || body.success === false) {
    const errorMsg = body?.error?.message || `Lỗi HTTP ${res.status}`;
    throw new Error(errorMsg);
  }

  return body.data as T;
}

export const api = {
  // Auth
  login: async (username: string, password: string) => {
    return request<{ user: UserClaims; token: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  },
  getMe: async () => request<UserClaims>('/auth/me'),
  logout: async () => request<{ message: string }>('/auth/logout', { method: 'POST' }),

  // Departments
  getDepartmentTree: async () => request<Department[]>('/departments?format=tree'),
  getDepartments: async () => request<Department[]>('/departments'),
  getDepartmentById: async (id: string) => request<Department>(`/departments/${id}`),
  createDepartment: async (code: string, name: string, parentId?: string, idempotencyKey?: string) => {
    const headers: Record<string, string> = {};
    if (idempotencyKey) headers['Idempotency-Key'] = idempotencyKey;
    return request<Department>('/departments', {
      method: 'POST',
      headers,
      body: JSON.stringify({ code, name, parentId: parentId || '' }),
    });
  },
  updateDepartment: async (id: string, name: string, version: number) => {
    return request<Department>(`/departments/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name, version }),
    });
  },
  moveDepartment: async (id: string, newParentId: string) => {
    return request<{ message: string }>(`/departments/${id}/move`, {
      method: 'POST',
      body: JSON.stringify({ newParentId }),
    });
  },
  archiveDepartment: async (id: string) => {
    return request<{ message: string }>(`/departments/${id}/archive`, {
      method: 'POST',
    });
  },
  assignDepartmentManager: async (id: string, employeeId: string) => {
    return request<{ message: string }>(`/departments/${id}/manager`, {
      method: 'POST',
      body: JSON.stringify({ employeeId }),
    });
  },

  // Employees
  getEmployees: async (departmentId?: string, status?: string) => {
    const params = new URLSearchParams();
    if (departmentId) params.append('departmentId', departmentId);
    if (status) params.append('status', status);
    const query = params.toString() ? `?${params.toString()}` : '';
    return request<Employee[]>(`/employees${query}`);
  },
  getEmployeeById: async (id: string) => request<Employee>(`/employees/${id}`),
  createEmployee: async (data: { code: string; fullName: string; email: string; departmentId: string; title: string; joinedAt: string }, idempotencyKey?: string) => {
    const headers: Record<string, string> = {};
    if (idempotencyKey) headers['Idempotency-Key'] = idempotencyKey;
    return request<Employee>('/employees', {
      method: 'POST',
      headers,
      body: JSON.stringify(data),
    });
  },
  transferEmployee: async (id: string, newDepartmentId: string, confirmManagerRemoval: boolean = false) => {
    return request<{ message: string }>(`/employees/${id}/transfer`, {
      method: 'POST',
      body: JSON.stringify({ newDepartmentId, confirmManagerRemoval }),
    });
  },
  resignEmployee: async (id: string, reason?: string) => {
    return request<{ message: string }>(`/employees/${id}/resign`, {
      method: 'POST',
      body: JSON.stringify({ reason: reason || 'Nghỉ việc theo nguyện vọng cá nhân' }),
    });
  },

  // Audit Logs
  getAuditLogs: async (targetId?: string) => {
    const query = targetId ? `?targetId=${encodeURIComponent(targetId)}` : '';
    return request<AuditLog[]>(`/audit-logs${query}`);
  },
};
