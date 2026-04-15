import { atom, map } from 'nanostores';

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  roles: string[];
  avatar_key?: string;
  avatar_url?: string;
  tenant_id?: string;
}

export interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
}

export const $auth = map<AuthState>({
  token: null,
  user: null,
  isAuthenticated: false,
});

function parseCookies(): Record<string, string> {
  return Object.fromEntries(
    document.cookie.split('; ').map(c => {
      const [key, ...rest] = c.split('=');
      return [key, rest.join('=')];
    })
  );
}

export function setAuth(token: string, user: User) {
  $auth.setKey('token', token);
  $auth.setKey('user', user);
  $auth.setKey('isAuthenticated', true);
  localStorage.setItem('oscar_auth', JSON.stringify({ token, user }));
}

export function updateToken(token: string) {
  $auth.setKey('token', token);
  const stored = localStorage.getItem('oscar_auth');
  if (stored) {
    try {
      const auth = JSON.parse(stored);
      localStorage.setItem('oscar_auth', JSON.stringify({ ...auth, token }));
    } catch {}
  }
}

export function clearAuth() {
  $auth.setKey('token', null);
  $auth.setKey('user', null);
  $auth.setKey('isAuthenticated', false);
  localStorage.removeItem('oscar_auth');
}

export function initAuth() {
  const stored = localStorage.getItem('oscar_auth');
  if (stored) {
    try {
      const { token, user } = JSON.parse(stored);
      if (token && user) {
        setAuth(token, user);
        return;
      }
    } catch {}
  }

  const cookies = parseCookies();
  const token = cookies['oscar_token'];
  const userStr = cookies['oscar_user'];
  
  if (token && userStr) {
    try {
      const user = JSON.parse(decodeURIComponent(userStr));
      setAuth(token, user);
    } catch {
      clearAuth();
    }
  } else {
    clearAuth();
  }
}

export function updateAuthUser(updates: Partial<User>) {
  const current = $auth.get();
  if (current.user) {
    const updatedUser = { ...current.user, ...updates };
    $auth.setKey('user', updatedUser);
    localStorage.setItem('oscar_auth', JSON.stringify({ token: current.token, user: updatedUser }));
  }
}

export function isOwnerOrAdmin(): boolean {
  const user = $auth.get().user;
  if (!user) return false;
  return user.roles?.includes('Owner') || user.roles?.includes('Admin') || false;
}

export function getToken(): string | null {
  const auth = $auth.get();
  if (auth.token) return auth.token;
  
  const cookies = parseCookies();
  return cookies['oscar_token'] || null;
}
