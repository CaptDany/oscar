const API_BASE_URL = 'http://localhost:8080/api/v1';

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

async function apiFetch<T>(endpoint: string, options: FetchOptions = {}): Promise<T> {
  const { token, ...fetchOptions } = options;
  
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
  });

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
  },

  users: {
    list: (token: string) =>
      apiFetch<{ data: any[]; total: number }>('/users', { token }),
    get: (token: string, id: string) =>
      apiFetch<any>(`/users/${id}`, { token }),
    updateRoles: (token: string, id: string, roleIds: string[]) =>
      apiFetch<any>(`/users/${id}/roles`, { method: 'PUT', body: JSON.stringify({ role_ids: roleIds }), token }),
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
      apiFetch<{ avatar_url: string }>('/upload/avatar/confirm', {
        method: 'POST',
        body: JSON.stringify({ object_key: objectKey }),
        token,
      }),
  },
};

export type { ApiError };
