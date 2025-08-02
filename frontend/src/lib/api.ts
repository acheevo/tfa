import { config } from '../config/app';

// Types for API responses
export interface User {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  email_verified: boolean;
  role: 'user' | 'admin';
  status: 'active' | 'inactive' | 'suspended';
  preferences: UserPreferences;
  avatar?: string;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface UserPreferences {
  theme?: string;
  language?: string;
  timezone?: string;
  notifications?: NotificationPrefs;
  privacy?: PrivacyPrefs;
  custom?: Record<string, any>;
}

export interface NotificationPrefs {
  email: boolean;
  sms: boolean;
  push: boolean;
}

export interface PrivacyPrefs {
  profile_visible: boolean;
  show_email: boolean;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface ApiError {
  error: string;
  details?: Record<string, string>;
}

export interface MessageResponse {
  message: string;
}

// Request types
export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
  confirm_password: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

export interface EmailVerificationRequest {
  token: string;
}

// User Management Types
export interface UpdateProfileRequest {
  first_name: string;
  last_name: string;
  avatar?: string;
}

export interface UpdatePreferencesRequest {
  theme?: string;
  language?: string;
  timezone?: string;
  notifications: NotificationPrefs;
  privacy: PrivacyPrefs;
  custom?: Record<string, any>;
}

export interface ChangeEmailRequest {
  new_email: string;
  password: string;
}

export interface UserSummary {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  role: 'user' | 'admin';
  status: 'active' | 'inactive' | 'suspended';
  email_verified: boolean;
  avatar?: string;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface UserListRequest {
  page?: number;
  page_size?: number;
  search?: string;
  role?: 'user' | 'admin';
  status?: 'active' | 'inactive' | 'suspended';
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface UserListResponse {
  users: UserSummary[];
  pagination: Pagination;
}

export interface Pagination {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface DashboardResponse {
  user: User;
  stats: UserStats;
  recent_logins: LoginHistoryEntry[];
  notifications: NotificationItem[];
}

export interface UserStats {
  total_logins: number;
  last_login_at: string | null;
  account_age_days: number;
  profile_complete: boolean;
}

export interface LoginHistoryEntry {
  id: number;
  ip_address: string;
  user_agent: string;
  success: boolean;
  created_at: string;
}

export interface NotificationItem {
  id: number;
  type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  message: string;
  read: boolean;
  priority: 'low' | 'medium' | 'high';
  created_at: string;
}

// Admin Types
export interface UpdateUserRoleRequest {
  role: 'user' | 'admin';
  reason: string;
}

export interface UpdateUserStatusRequest {
  status: 'active' | 'inactive' | 'suspended';
  reason: string;
}

export interface AdminUpdateUserRequest {
  first_name?: string;
  last_name?: string;
  email?: string;
  email_verified?: boolean;
  role?: 'user' | 'admin';
  status?: 'active' | 'inactive' | 'suspended';
  avatar?: string;
  reason: string;
}

export interface DeleteUserRequest {
  reason: string;
  force?: boolean;
}

export interface BulkUserActionRequest {
  user_ids: number[];
  action: 'activate' | 'deactivate' | 'suspend' | 'delete' | 'role_change';
  role?: 'user' | 'admin';
  reason: string;
}

export interface BulkActionResult {
  total_requested: number;
  successful: number;
  failed: number;
  results: BulkActionItemResult[];
}

export interface BulkActionItemResult {
  user_id: number;
  success: boolean;
  error?: string;
}

export interface AdminStatsResponse {
  total_users: number;
  active_users: number;
  inactive_users: number;
  suspended_users: number;
  admin_users: number;
  new_users_today: number;
  new_users_this_week: number;
  user_growth: UserGrowthData[];
  top_countries?: CountryData[];
}

export interface UserGrowthData {
  date: string;
  count: number;
}

export interface CountryData {
  country: string;
  count: number;
}

export interface AdminAuditLogRequest {
  page?: number;
  page_size?: number;
  user_id?: number;
  target_id?: number;
  action?: string;
  level?: 'info' | 'warning' | 'error';
  resource?: string;
  date_from?: string;
  date_to?: string;
  ip_address?: string;
}

export interface AdminAuditLogResponse {
  logs: EnhancedAuditLogEntry[];
  pagination: Pagination;
}

export interface EnhancedAuditLogEntry {
  id: number;
  action: string;
  level: 'info' | 'warning' | 'error';
  resource: string;
  description: string;
  ip_address: string;
  user_agent: string;
  metadata?: Record<string, any>;
  created_at: string;
  user?: UserSummary;
  target?: UserSummary;
}

export interface UserDetailResponse extends User {
  login_history?: LoginHistoryEntry[];
  audit_trail?: AuditLogEntry[];
}

export interface AuditLogEntry {
  id: number;
  action: string;
  level: 'info' | 'warning' | 'error';
  resource: string;
  description: string;
  ip_address: string;
  user_agent: string;
  metadata?: Record<string, any>;
  created_at: string;
}

// API Client class
class ApiClient {
  private baseURL: string;
  private refreshPromise: Promise<void> | null = null;

  constructor() {
    this.baseURL = config.apiUrl;
  }

  // Generic request method
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const defaultHeaders: HeadersInit = {
      'Content-Type': 'application/json',
    };

    const config: RequestInit = {
      credentials: 'include', // Include cookies for authentication
      ...options,
      headers: {
        ...defaultHeaders,
        ...options.headers,
      },
    };

    let response = await fetch(url, config);

    // Handle token refresh on 401 errors
    if (response.status === 401 && endpoint !== '/auth/refresh' && endpoint !== '/auth/login') {
      const refreshed = await this.refreshToken();
      if (refreshed) {
        // Retry the request after successful refresh
        response = await fetch(url, config);
      }
    }

    const contentType = response.headers.get('content-type');
    const isJson = contentType?.includes('application/json');

    if (!response.ok) {
      let errorData: ApiError;
      
      if (isJson) {
        errorData = await response.json();
      } else {
        errorData = {
          error: `HTTP ${response.status}: ${response.statusText}`,
          name: 'HTTPError',
          message: `HTTP ${response.status}: ${response.statusText}`,
        };
      }

      throw new ApiError(errorData.error, response.status, errorData.details);
    }

    if (isJson) {
      return response.json();
    }
    
    return response.text() as unknown as T;
  }

  // Authentication methods
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async logout(): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/logout', {
      method: 'POST',
    });
  }

  async logoutAll(): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/logout-all', {
      method: 'POST',
    });
  }

  async refreshToken(): Promise<boolean> {
    // Prevent multiple refresh attempts
    if (this.refreshPromise) {
      await this.refreshPromise;
      return true;
    }

    this.refreshPromise = this.performRefresh();
    
    try {
      await this.refreshPromise;
      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      return false;
    } finally {
      this.refreshPromise = null;
    }
  }

  private async performRefresh(): Promise<void> {
    await this.request<AuthResponse>('/auth/refresh', {
      method: 'POST',
    });
  }

  async checkAuth(): Promise<{ authenticated: boolean; user?: User }> {
    return this.request<{ authenticated: boolean; user?: User }>('/auth/check');
  }

  async getProfile(): Promise<User> {
    return this.request<User>('/auth/profile');
  }

  async verifyEmail(data: EmailVerificationRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/verify-email', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async resendEmailVerification(): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/resend-verification', {
      method: 'POST',
    });
  }

  async forgotPassword(data: ForgotPasswordRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async resetPassword(data: ResetPasswordRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async changePassword(data: ChangePasswordRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // User Management methods
  async getUserProfile(): Promise<User> {
    return this.request<User>('/user/profile');
  }

  async updateProfile(data: UpdateProfileRequest): Promise<User> {
    return this.request<User>('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async getPreferences(): Promise<UserPreferences> {
    return this.request<UserPreferences>('/user/preferences');
  }

  async updatePreferences(data: UpdatePreferencesRequest): Promise<UserPreferences> {
    return this.request<UserPreferences>('/user/preferences', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async changeEmail(data: ChangeEmailRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>('/user/change-email', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getDashboard(): Promise<DashboardResponse> {
    return this.request<DashboardResponse>('/user/dashboard');
  }

  // Admin methods
  async getUsers(params?: UserListRequest): Promise<UserListResponse> {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
    if (params?.search) searchParams.append('search', params.search);
    if (params?.role) searchParams.append('role', params.role);
    if (params?.status) searchParams.append('status', params.status);
    if (params?.sort_by) searchParams.append('sort_by', params.sort_by);
    if (params?.sort_order) searchParams.append('sort_order', params.sort_order);

    const queryString = searchParams.toString();
    const endpoint = queryString ? `/admin/users?${queryString}` : '/admin/users';
    
    return this.request<UserListResponse>(endpoint);
  }

  async getUserDetails(userId: number): Promise<UserDetailResponse> {
    return this.request<UserDetailResponse>(`/admin/users/${userId}`);
  }

  async updateUser(userId: number, data: AdminUpdateUserRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>(`/admin/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async updateUserRole(userId: number, data: UpdateUserRoleRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>(`/admin/users/${userId}/role`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async updateUserStatus(userId: number, data: UpdateUserStatusRequest): Promise<MessageResponse> {
    return this.request<MessageResponse>(`/admin/users/${userId}/status`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteUsers(userIds: number[], data: DeleteUserRequest): Promise<MessageResponse> {
    const idsParam = userIds.join(',');
    return this.request<MessageResponse>(`/admin/users?ids=${idsParam}`, {
      method: 'DELETE',
      body: JSON.stringify(data),
    });
  }

  async bulkUpdateUsers(data: BulkUserActionRequest): Promise<BulkActionResult> {
    return this.request<BulkActionResult>('/admin/users/bulk', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getAdminStats(): Promise<AdminStatsResponse> {
    return this.request<AdminStatsResponse>('/admin/stats');
  }

  async getAuditLogs(params?: AdminAuditLogRequest): Promise<AdminAuditLogResponse> {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
    if (params?.user_id) searchParams.append('user_id', params.user_id.toString());
    if (params?.target_id) searchParams.append('target_id', params.target_id.toString());
    if (params?.action) searchParams.append('action', params.action);
    if (params?.level) searchParams.append('level', params.level);
    if (params?.resource) searchParams.append('resource', params.resource);
    if (params?.date_from) searchParams.append('date_from', params.date_from);
    if (params?.date_to) searchParams.append('date_to', params.date_to);
    if (params?.ip_address) searchParams.append('ip_address', params.ip_address);

    const queryString = searchParams.toString();
    const endpoint = queryString ? `/admin/audit-logs?${queryString}` : '/admin/audit-logs';
    
    return this.request<AdminAuditLogResponse>(endpoint);
  }

  // Health check
  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    return this.request<{ status: string; timestamp: string }>('/health');
  }
}

// Custom error class
export class ApiError extends Error {
  constructor(
    message: string,
    public status?: number,
    public details?: Record<string, string>
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// Export singleton instance
export const api = new ApiClient();