interface User {
  id: string;
  tenant_id?: string;
  email: string;
  first_name: string;
  last_name: string;
  roles: string[];
}

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
}

const STORAGE_KEY = 'oscar_auth';

function parseCookies(): Record<string, string> {
  if (typeof window === 'undefined') return {};
  return Object.fromEntries(
    document.cookie.split('; ').map(c => {
      const [key, ...rest] = c.split('=');
      return [key, rest.join('=')];
    })
  );
}

class AuthStore {
  private state: AuthState;

  constructor() {
    this.state = this.loadFromStorage();
  }

  private loadFromStorage(): AuthState {
    if (typeof window === 'undefined') {
      return { token: null, refreshToken: null, user: null, isAuthenticated: false };
    }
    
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        return {
          ...parsed,
          isAuthenticated: !!parsed.token,
        };
      } catch {
        return { token: null, refreshToken: null, user: null, isAuthenticated: false };
      }
    }
    
    const cookies = parseCookies();
    const token = cookies['oscar_token'];
    const refreshToken = cookies['oscar_refresh_token'];
    const userStr = cookies['oscar_user'];
    
    if (token && userStr) {
      try {
        const user = JSON.parse(decodeURIComponent(userStr));
        return {
          token,
          refreshToken: refreshToken || null,
          user,
          isAuthenticated: true,
        };
      } catch {}
    }
    
    return { token: null, refreshToken: null, user: null, isAuthenticated: false };
  }

  private saveToStorage(): void {
    if (typeof window === 'undefined') return;
    
    localStorage.setItem(STORAGE_KEY, JSON.stringify({
      token: this.state.token,
      refreshToken: this.state.refreshToken,
      user: this.state.user,
    }));
  }

  get token(): string | null {
    return this.state.token;
  }

  get refreshToken(): string | null {
    return this.state.refreshToken;
  }

  get user(): User | null {
    return this.state.user;
  }

  get isAuthenticated(): boolean {
    return this.state.isAuthenticated;
  }

  setAuth(token: string, user: User, refreshToken?: string): void {
    this.state = {
      token,
      refreshToken: refreshToken || null,
      user,
      isAuthenticated: true,
    };
    this.saveToStorage();
  }

  updateToken(token: string): void {
    this.state.token = token;
    this.saveToStorage();
  }

  updateUser(user: User): void {
    this.state.user = user;
    this.saveToStorage();
  }

  clear(): void {
    this.state = {
      token: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,
    };
    if (typeof window !== 'undefined') {
      localStorage.removeItem(STORAGE_KEY);
    }
  }
}

export const auth = new AuthStore();

let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback);
}

function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach(callback => callback(newToken));
  refreshSubscribers = [];
}

export async function refreshAccessToken(): Promise<string | null> {
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
      clearSession();
      window.location.href = '/login?reason=session_expired';
      return null;
    }

    const data = await res.json();
    
    if (data.success && data.token) {
      auth.updateToken(data.token);
      onTokenRefreshed(data.token);
      return data.token;
    }

    clearSession();
    window.location.href = '/login?reason=session_expired';
    return null;
  } catch (error) {
    console.error('Token refresh failed:', error);
    clearSession();
    window.location.href = '/login?reason=session_expired';
    return null;
  } finally {
    isRefreshing = false;
  }
}

export async function authenticatedFetch(url: string, options: RequestInit = {}): Promise<Response> {
  let token = auth.token;
  
  const headers = {
    ...(options.headers || {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  let response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (response.status === 401) {
    token = await refreshAccessToken();
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
      response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include',
      });
    }
  }

  return response;
}

export function clearSession(): void {
  auth.clear();
  document.cookie = 'oscar_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
  document.cookie = 'oscar_refresh_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
  document.cookie = 'oscar_user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
}

export function logout(): void {
  clearSession();
  window.location.href = '/login';
}

export function getToken(): string | null {
  return auth.token;
}
