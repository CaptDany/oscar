interface User {
  id: string;
  tenant_id: string;
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

  setAuth(token: string, refreshToken: string, user: User): void {
    this.state = {
      token,
      refreshToken,
      user,
      isAuthenticated: true,
    };
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
