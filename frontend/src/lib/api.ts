// Override with NEXT_PUBLIC_API_URL when running against another environment.
const API_BASE = (process.env.NEXT_PUBLIC_API_URL || 'https://4lh3vbgsp1.execute-api.ap-southeast-1.amazonaws.com/api/v1').replace(/\/$/, '');

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
  directEmployeeCount: number;
  createdAt: string;
  updatedAt: string;
  children?: Department[];
}

export interface Employee {
  id: string;
  code: string;
  fullName: string;
  avatar?: string;
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
  targetName?: string;
  before?: Record<string, any>;
  after?: Record<string, any>;
  occurredAt: string;
}

// Chỉ lưu thông tin giao diện; token xác thực nằm trong cookie HttpOnly.
export function setStoredUser(user: UserClaims) {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('minihrm_token');
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

export class ApiError extends Error {
  constructor(message: string, readonly status: number) {
    super(message);
    this.name = "ApiError";
  }
}

export function apiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError && (error.status === 403 || error.status === 404)) {
    return "Không tìm thấy dữ liệu hoặc bạn không có quyền truy cập.";
  }
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

const auditActionLabels: Record<string, string> = {
  "department.created": "Tạo phòng ban",
  "department.updated": "Cập nhật phòng ban",
  "department.renamed": "Đổi tên phòng ban",
  "department.moved": "Di chuyển phòng ban",
  "department.archived": "Lưu trữ phòng ban",
  "department.manager_assigned": "Bổ nhiệm trưởng phòng",
  "employee.created": "Tạo hồ sơ nhân viên",
  "employee.updated": "Cập nhật hồ sơ nhân viên",
  "employee.transferred": "Chuyển phòng ban nhân viên",
  "employee.resigned": "Cho nhân viên nghỉ việc",
  "employee.avatar_updated": "Cập nhật ảnh đại diện",
  "employee.avatar_removed": "Xóa ảnh đại diện",
  "system.seeded": "Khởi tạo dữ liệu mẫu",
};

export function getAuditActionLabel(action: string): string {
  return auditActionLabels[action] || "Thay đổi dữ liệu hệ thống";
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };
	if (options.method?.toUpperCase() === 'POST' && !path.startsWith('/auth/') && !headers['Idempotency-Key']) {
		headers['Idempotency-Key'] = crypto.randomUUID();
	}

  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, { ...options, headers, credentials: 'include' });
  } catch {
    throw new ApiError("Mất kết nối tới máy chủ. Vui lòng kiểm tra đường truyền rồi thử lại.", 0);
  }

  const body = await res.json();
  if (!res.ok || body.success === false) {
    const errorMsg = body?.error?.message || `Lỗi HTTP ${res.status}`;
    throw new ApiError(errorMsg, res.status);
  }

  return body.data as T;
}

export const api = {
  // Auth
  login: async (username: string, password: string) => {
    return request<{ user: UserClaims }>('/auth/login', {
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
  getEmployees: async (departmentId?: string, status?: string, search?: string, includeDescendants = false) => {
    const params = new URLSearchParams();
    if (departmentId) params.append('departmentId', departmentId);
    if (status) params.append('status', status);
    if (search) params.append('search', search);
    if (departmentId && includeDescendants) params.append('scope', 'subtree');
    const query = params.toString() ? `?${params.toString()}` : '';
    return request<Employee[]>(`/employees${query}`);
  },
  getEmployeeById: async (id: string) => request<Employee>(`/employees/${id}`),
  updateEmployee: async (id: string, data: { fullName: string; title: string; joinedAt: string; version: number; avatar?: string }) => request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
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
