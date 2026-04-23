const API_BASE_URL = '/api/v1';

export { auth } from './auth';
export type { ApiError };

interface ApiError {
  code: string;
  message: string;
  details?: any;
}

interface FetchOptions extends RequestInit {
  token?: string;
}

let isAuthFailed = false;
let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback);
}

function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach(callback => callback(newToken));
  refreshSubscribers = [];
}

function getStoredToken(): string | null {
  const stored = localStorage.getItem('oscar_auth');
  if (stored) {
    try {
      const { token } = JSON.parse(stored);
      if (token) return token;
    } catch {}
  }
  const cookies = Object.fromEntries(
    document.cookie.split('; ').map(c => c.split('='))
  );
  return cookies['oscar_token'] || null;
}

function updateStoredToken(token: string): void {
  const stored = localStorage.getItem('oscar_auth');
  const current = stored ? JSON.parse(stored) : {};
  localStorage.setItem('oscar_auth', JSON.stringify({ ...current, token }));
}

function clearStoredSession(): void {
  localStorage.removeItem('oscar_auth');
  document.cookie = 'oscar_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
  document.cookie = 'oscar_refresh_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
  document.cookie = 'oscar_user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
}

async function refreshAccessToken(): Promise<string | null> {
  if (isAuthFailed) return null;
  
  if (isRefreshing) {
    return new Promise(resolve => {
      subscribeTokenRefresh((token) => resolve(token));
    });
  }

  isRefreshing = true;

  try {
    const res = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include',
    });

    if (!res.ok) {
      clearStoredSession();
      window.location.href = '/login?reason=session_expired';
      return null;
    }

    const data = await res.json();
    
    if (data.success && data.token) {
      updateStoredToken(data.token);
      onTokenRefreshed(data.token);
      return data.token;
    }

    clearStoredSession();
    window.location.href = '/login?reason=session_expired';
    return null;
  } catch (error) {
    console.error('Token refresh failed:', error);
    clearStoredSession();
    window.location.href = '/login?reason=session_expired';
    return null;
  } finally {
    isRefreshing = false;
  }
}

export function resetApiAuthState(): void {
  isAuthFailed = false;
  isRefreshing = false;
  refreshSubscribers = [];
}

async function apiFetch<T>(endpoint: string, options: FetchOptions = {}): Promise<T> {
  if (isAuthFailed) {
    throw { code: 'AUTH_FAILED', message: 'Session expired', details: null };
  }

  const token = options.token || getStoredToken();
  const { token: _, ...fetchOptions } = options;
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...fetchOptions,
    headers,
    credentials: 'include',
  });

  if (response.status === 401) {
    isAuthFailed = true;
    const newToken = await refreshAccessToken();
    
    if (newToken) {
      isAuthFailed = false;
      (headers as Record<string, string>)['Authorization'] = `Bearer ${newToken}`;
      const retryResponse = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...fetchOptions,
        headers,
        credentials: 'include',
      });

      if (!retryResponse.ok) {
        const errorData = await retryResponse.json().catch(() => ({}));
        const error: ApiError = {
          code: errorData.error?.code || 'UNKNOWN_ERROR',
          message: errorData.error?.message || retryResponse.statusText,
          details: errorData.error?.details,
        };
        throw error;
      }

      return retryResponse.json();
    }
    
    throw { code: 'AUTH_FAILED', message: 'Session expired', details: null };
  }

  if (response.status === 429) {
    const retryAfter = response.headers.get('retry-after');
    const waitMs = retryAfter ? parseInt(retryAfter) * 1000 : 5000;
    console.warn(`Rate limited. Waiting ${waitMs}ms...`);
    await new Promise(resolve => setTimeout(resolve, waitMs));
    return apiFetch(endpoint, options);
  }

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    const error: ApiError = {
      code: errorData.error?.code || 'UNKNOWN_ERROR',
      message: errorData.error?.message || response.statusText,
      details: errorData.error?.details,
    };
    throw error;
  }

  return response.json();
}

export const api = {
  auth: {
    login: (email: string, password: string) =>
      apiFetch<{ access_token: string; refresh_token: string; user: any }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    register: (data: { email: string; password: string; first_name: string; last_name: string; tenant_name: string; tenant_slug: string }) =>
      apiFetch<{ message: string }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    me: (token: string) =>
      apiFetch<any>('/auth/me', { token }),
    logout: (token: string) =>
      apiFetch<{ message: string }>('/auth/logout', { method: 'POST', token }),
  },

  persons: {
    list: (token: string, params?: { limit?: number; cursor?: string; search?: string; type?: string }) => {
      const searchParams = new URLSearchParams();
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      if (params?.cursor) searchParams.set('cursor', params.cursor);
      if (params?.search) searchParams.set('search', params.search);
      if (params?.type) searchParams.set('type', params.type);
      const query = searchParams.toString();
      return apiFetch<{ data: any[]; total: number; next_cursor: string | null }>(`/persons${query ? `?${query}` : ''}`, { token });
    },
    get: (token: string, id: string) =>
      apiFetch<any>(`/persons/${id}`, { token }),
    create: (token: string, data: any) =>
      apiFetch<any>('/persons', { method: 'POST', body: JSON.stringify(data), token }),
    update: (token: string, id: string, data: any) =>
      apiFetch<any>(`/persons/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/persons/${id}`, { method: 'DELETE', token }),
  },

  companies: {
    list: (token: string, params?: { limit?: number; offset?: number; search?: string; includeTotal?: boolean }) => {
      const searchParams = new URLSearchParams();
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      if (params?.offset) searchParams.set('offset', params.offset.toString());
      if (params?.search) searchParams.set('search', params.search);
      if (params?.includeTotal) searchParams.set('include_total', 'true');
      const query = searchParams.toString();
      return apiFetch<{ data: any[]; total: number }>(`/companies${query ? `?${query}` : ''}`, { token });
    },
    get: (token: string, id: string) =>
      apiFetch<any>(`/companies/${id}`, { token }),
    create: (token: string, data: any) =>
      apiFetch<any>('/companies', { method: 'POST', body: JSON.stringify(data), token }),
    update: (token: string, id: string, data: any) =>
      apiFetch<any>(`/companies/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/companies/${id}`, { method: 'DELETE', token }),
  },

  deals: {
    list: (token: string, params?: { limit?: number; offset?: number; search?: string }) => {
      const searchParams = new URLSearchParams();
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      if (params?.offset) searchParams.set('offset', params.offset.toString());
      if (params?.search) searchParams.set('search', params.search);
      const query = searchParams.toString();
      return apiFetch<{ data: any[]; total: number }>(`/deals${query ? `?${query}` : ''}`, { token });
    },
    kanban: (token: string) =>
      apiFetch<any[]>('/deals/kanban', { token }),
    get: (token: string, id: string) =>
      apiFetch<any>(`/deals/${id}`, { token }),
    create: (token: string, data: any) =>
      apiFetch<any>('/deals', { method: 'POST', body: JSON.stringify(data), token }),
    update: (token: string, id: string, data: any) =>
      apiFetch<any>(`/deals/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/deals/${id}`, { method: 'DELETE', token }),
  },

  activities: {
    list: (token: string, params?: { limit?: number; offset?: number; type?: string }) => {
      const searchParams = new URLSearchParams();
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      if (params?.offset) searchParams.set('offset', params.offset.toString());
      if (params?.type) searchParams.set('type', params.type);
      const query = searchParams.toString();
      return apiFetch<{ data: any[]; total: number }>(`/activities${query ? `?${query}` : ''}`, { token });
    },
    get: (token: string, id: string) =>
      apiFetch<any>(`/activities/${id}`, { token }),
    create: (token: string, data: any) =>
      apiFetch<any>('/activities', { method: 'POST', body: JSON.stringify(data), token }),
    complete: (token: string, id: string) =>
      apiFetch<any>(`/activities/${id}/complete`, { method: 'POST', token }),
    uncomplete: (token: string, id: string) =>
      apiFetch<any>(`/activities/${id}/uncomplete`, { method: 'POST', token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/activities/${id}`, { method: 'DELETE', token }),
  },

  pipelines: {
    list: (token: string) =>
      apiFetch<{ data: any[] }>('/pipelines', { token }),
    get: (token: string, id: string) =>
      apiFetch<any>(`/pipelines/${id}`, { token }),
    getStages: (token: string, id: string) =>
      apiFetch<any[]>(`/pipelines/${id}/stages`, { token }),
    create: (token: string, data: { name: string; currency?: string; is_default?: boolean }) =>
      apiFetch<any>('/pipelines', { method: 'POST', body: JSON.stringify(data), token }),
    update: (token: string, id: string, data: { name?: string; currency?: string; is_default?: boolean }) =>
      apiFetch<any>(`/pipelines/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/pipelines/${id}`, { method: 'DELETE', token }),
    createStage: (token: string, pipelineId: string, data: { name: string; probability?: number; stage_type?: string }) =>
      apiFetch<any>(`/pipelines/${pipelineId}/stages`, { method: 'POST', body: JSON.stringify(data), token }),
    updateStage: (token: string, pipelineId: string, stageId: string, data: { name?: string; probability?: number; stage_type?: string }) =>
      apiFetch<any>(`/pipelines/${pipelineId}/stages/${stageId}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    deleteStage: (token: string, pipelineId: string, stageId: string) =>
      apiFetch<any>(`/pipelines/${pipelineId}/stages/${stageId}`, { method: 'DELETE', token }),
    reorderStages: (token: string, pipelineId: string, stageIds: string[]) =>
      apiFetch<any>(`/pipelines/${pipelineId}/stages/reorder`, { method: 'PATCH', body: JSON.stringify({ stage_ids: stageIds }), token }),
  },

  users: {
    list: (token: string) =>
      apiFetch<{ data: any[]; total: number }>('/users', { token }),
    get: (token: string, id: string) =>
      apiFetch<any>(`/users/${id}`, { token }),
    update: (token: string, id: string, data: { first_name?: string; last_name?: string; avatar_url?: string; timezone?: string; locale?: string; is_active?: boolean }) =>
      apiFetch<any>(`/users/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    updateRoles: (token: string, id: string, roleIds: string[]) =>
      apiFetch<any>(`/users/${id}/roles`, { method: 'PUT', body: JSON.stringify({ role_ids: roleIds }), token }),
  },

  notifications: {
    list: (token: string, params?: { limit?: number; cursor?: string; unread_only?: boolean }) => {
      const searchParams = new URLSearchParams();
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      if (params?.cursor) searchParams.set('cursor', params.cursor);
      if (params?.unread_only) searchParams.set('unread_only', 'true');
      const query = searchParams.toString();
      return apiFetch<{ data: any[]; total: number; next_cursor: string | null }>(`/notifications${query ? `?${query}` : ''}`, { token });
    },
    get: (token: string, id: string) =>
      apiFetch<any>(`/notifications/${id}`, { token }),
    count: (token: string) =>
      apiFetch<{ unread_count: number }>('/notifications/count', { token }),
    markAsRead: (token: string, id: string) =>
      apiFetch<any>(`/notifications/${id}/read`, { method: 'POST', token }),
    markAllAsRead: (token: string) =>
      apiFetch<{ marked_count: number }>('/notifications/read-all', { method: 'POST', token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/notifications/${id}`, { method: 'DELETE', token }),
  },

  teams: {
    list: (token: string, params?: { include_members?: boolean }) => {
      const searchParams = new URLSearchParams();
      if (params?.include_members) searchParams.set('include_members', 'true');
      const query = searchParams.toString();
      return apiFetch<{ data: any[] }>(`/teams${query ? `?${query}` : ''}`, { token });
    },
    get: (token: string, id: string) =>
      apiFetch<{ team: any; members: any[] }>(`/teams/${id}`, { token }),
    create: (token: string, data: { name: string; description?: string }) =>
      apiFetch<any>('/teams', { method: 'POST', body: JSON.stringify(data), token }),
    update: (token: string, id: string, data: { name?: string; description?: string }) =>
      apiFetch<any>(`/teams/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/teams/${id}`, { method: 'DELETE', token }),
    listMembers: (token: string, id: string) =>
      apiFetch<{ members: any[] }>(`/teams/${id}/members`, { token }),
    addMember: (token: string, teamId: string, userId: string, isLead?: boolean) =>
      apiFetch<any>(`/teams/${teamId}/members`, { method: 'POST', body: JSON.stringify({ user_id: userId, is_lead: isLead }), token }),
    removeMember: (token: string, teamId: string, userId: string) =>
      apiFetch<any>(`/teams/${teamId}/members/${userId}`, { method: 'DELETE', token }),
    setLead: (token: string, teamId: string, userId: string) =>
      apiFetch<any>(`/teams/${teamId}/lead/${userId}`, { method: 'POST', token }),
  },

  products: {
    list: (token: string, params?: { limit?: number; offset?: number; active_only?: boolean }) => {
      const searchParams = new URLSearchParams();
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      if (params?.offset) searchParams.set('offset', params.offset.toString());
      if (params?.active_only) searchParams.set('active_only', 'true');
      const query = searchParams.toString();
      return apiFetch<{ data: any[]; total: number }>(`/products${query ? `?${query}` : ''}`, { token });
    },
    get: (token: string, id: string) =>
      apiFetch<any>(`/products/${id}`, { token }),
    create: (token: string, data: { name: string; description?: string; sku?: string; price: number; currency?: string; unit?: string; is_active?: boolean }) =>
      apiFetch<any>('/products', { method: 'POST', body: JSON.stringify(data), token }),
    update: (token: string, id: string, data: { name?: string; description?: string; sku?: string; price?: number; currency?: string; unit?: string; is_active?: boolean }) =>
      apiFetch<any>(`/products/${id}`, { method: 'PATCH', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<any>(`/products/${id}`, { method: 'DELETE', token }),
  },

  upload: {
    getAvatarPresignedURL: (token: string, filename: string, contentType: string) =>
      apiFetch<{ upload_url: string; object_key: string; final_url: string }>('/upload/avatar', {
        method: 'POST',
        body: JSON.stringify({ filename, content_type: contentType }),
        token,
      }),
    uploadToPresignedURL: async (presignedUrl: string, file: File): Promise<void> => {
      const response = await fetch(presignedUrl, {
        method: 'PUT',
        body: file,
        headers: {
          'Content-Type': file.type,
        },
      });
      if (!response.ok) {
        throw new Error('Failed to upload file');
      }
    },
    confirmAvatarUpload: (token: string, objectKey: string) =>
      apiFetch<{ avatar_key: string }>('/upload/avatar/confirm', {
        method: 'POST',
        body: JSON.stringify({ object_key: objectKey }),
        token,
      }),
    getAvatarURL: (userId: string) => `/api/v1/avatar/${userId}`,
  },

  settings: {
    get: (token: string) =>
      apiFetch<{ data: any }>('/settings', { token }),
    update: (token: string, data: { name?: string; currency?: string; timezone?: string }) =>
      apiFetch<{ message: string }>('/settings', { method: 'PATCH', body: JSON.stringify(data), token }),
  },

  invitations: {
    list: (token: string) =>
      apiFetch<{ data: any[]; total: number }>('/invitations', { token }),
    create: (token: string, data: { email: string; first_name: string; last_name: string; role_name: string }) =>
      apiFetch<{ data: any }>('/invitations', { method: 'POST', body: JSON.stringify(data), token }),
    delete: (token: string, id: string) =>
      apiFetch<{ message: string }>(`/invitations/${id}`, { method: 'DELETE', token }),
  },
};

export type { ApiError };
