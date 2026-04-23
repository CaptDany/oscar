import { useState, useEffect } from 'preact/hooks';

interface User {
  id: string;
  first_name?: string;
  last_name?: string;
  email?: string;
  avatar_key?: string;
}

interface Notification {
  id: string;
  title: string;
  body?: string;
  message?: string;
  read_at?: string;
  created_at: string;
}

export function UserMenu() {
  const [open, setOpen] = useState(false);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const cookies = parseCookies();
    const token = cookies['oscar_token'];
    const userStr = cookies['oscar_user'];

    if (token && userStr) {
      try {
        setUser(JSON.parse(decodeURIComponent(userStr)));
      } catch (e) {}
    }

    const stored = localStorage.getItem('oscar_auth');
    if (stored) {
      try {
        const { user: storedUser } = JSON.parse(stored);
        if (storedUser) setUser(storedUser);
      } catch (e) {}
    }
  }, []);

  const initials = `${user?.first_name?.[0] || ''}${user?.last_name?.[0] || ''}`.toUpperCase();
  const fullName = `${user?.first_name || ''} ${user?.last_name || ''}`.trim();

  const handleLogout = async () => {
    const token = getToken();
    if (token) {
      try {
        await fetch('/api/v1/auth/logout', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` }
        });
      } catch (e) {}
    }
    localStorage.removeItem('oscar_auth');
    document.cookie = 'oscar_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    document.cookie = 'oscar_refresh_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    document.cookie = 'oscar_user=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    window.location.href = '/login';
  };

  return (
    <div class="relative" id="user-dropdown-container">
      <button
        id="user-menu-btn"
        onClick={() => setOpen(!open)}
        class="flex items-center gap-3 hover:bg-surface-container-low rounded-lg px-2 py-1.5 transition-colors"
      >
        <div class="text-right hidden md:block">
          <p class="text-[11px] font-bold text-on-surface leading-tight">{fullName || 'User'}</p>
          <p class="text-[9px] text-gray-500">USER</p>
        </div>
        <div class="w-9 h-9 rounded-full ring-1 ring-primary/20 flex items-center justify-center bg-surface-container-highest">
          <span class="text-xs font-bold text-gray-400">{initials || '--'}</span>
        </div>
        <span class="material-symbols-outlined text-sm text-gray-500">expand_more</span>
      </button>

      {open && (
        <div class="absolute right-0 top-full mt-2 w-56 bg-surface-container border border-outline-variant/20 rounded-xl shadow-2xl overflow-hidden z-50">
          <div class="px-4 py-3 border-b border-outline-variant/10">
            <p class="text-sm font-semibold">{fullName || 'User'}</p>
            <p class="text-[10px] text-gray-500 truncate">{user?.email || ''}</p>
          </div>
          <div class="py-1">
            <a href="/settings" class="flex items-center gap-3 px-4 py-2.5 hover:bg-surface-container-low transition-colors text-sm">
              <span class="material-symbols-outlined text-lg text-gray-400">settings</span>
              Settings
            </a>
            <a href="/teams" class="flex items-center gap-3 px-4 py-2.5 hover:bg-surface-container-low transition-colors text-sm">
              <span class="material-symbols-outlined text-lg text-gray-400">group</span>
              Team
            </a>
          </div>
          <div class="py-1 border-t border-outline-variant/10">
            <button
              onClick={handleLogout}
              class="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-error/10 transition-colors text-sm text-error"
            >
              <span class="material-symbols-outlined text-lg">logout</span>
              Sign out
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export function Notifications() {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [hasUnread, setHasUnread] = useState(false);

  useEffect(() => {
    loadNotifications();
    const interval = setInterval(loadNotifications, 60000);
    return () => clearInterval(interval);
  }, []);

  const loadNotifications = async () => {
    const token = getToken();
    if (!token) return;

    try {
      const res = await fetch('/api/v1/notifications?limit=10', {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (res.status === 401) {
        window.location.href = '/login?reason=session_expired';
        return;
      }

      if (res.status === 429) return;

      const data = await res.json();
      const notifs = data.data || [];
      setNotifications(notifs);
      setHasUnread(notifs.some((n: Notification) => !n.read_at));
    } catch (e) {}
  };

  const markAllRead = async () => {
    const token = getToken();
    if (!token) return;

    try {
      await fetch('/api/v1/notifications/read-all', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      setHasUnread(false);
      loadNotifications();
    } catch (e) {}
  };

  const formatTimeAgo = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const mins = Math.floor(diff / 60000);
    const hours = Math.floor(mins / 60);
    const days = Math.floor(hours / 24);
    if (mins < 1) return 'Just now';
    if (mins < 60) return `${mins}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div class="relative" id="notifications-dropdown-container">
      <button
        id="notifications-btn"
        onClick={() => setOpen(!open)}
        class="p-2.5 hover:bg-surface-container-low rounded-lg transition-colors active:opacity-80 relative text-gray-400 hover:text-primary"
      >
        <span class="material-symbols-outlined">notifications</span>
        {hasUnread && (
          <span class="absolute top-1.5 right-1.5 w-2.5 h-2.5 bg-primary rounded-full animate-pulse" />
        )}
      </button>

      {open && (
        <div class="absolute right-0 top-full mt-2 w-80 bg-surface-container border border-outline-variant/20 rounded-xl shadow-2xl overflow-hidden z-50">
          <div class="px-4 py-3 border-b border-outline-variant/10 flex items-center justify-between">
            <h3 class="font-semibold text-sm">Notifications</h3>
            {hasUnread && (
              <button onClick={markAllRead} class="text-[10px] text-primary hover:underline">
                Mark all read
              </button>
            )}
          </div>
          <div class="max-h-80 overflow-y-auto">
            {notifications.length === 0 ? (
              <div class="p-4 text-center text-gray-500 text-sm">
                <span class="material-symbols-outlined text-2xl">notifications_off</span>
                <p class="mt-2">No notifications</p>
              </div>
            ) : (
              notifications.map((n) => (
                <div
                  key={n.id}
                  class={`px-4 py-3 hover:bg-surface-container-low transition-colors border-b border-outline-variant/5 ${!n.read_at ? 'bg-primary/5' : ''}`}
                >
                  <p class="text-sm font-medium">{n.title || 'Notification'}</p>
                  <p class="text-xs text-gray-500 mt-0.5 line-clamp-2">{n.body || n.message || ''}</p>
                  <p class="text-[10px] text-gray-600 mt-1">{formatTimeAgo(n.created_at)}</p>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function parseCookies(): Record<string, string> {
  return Object.fromEntries(
    document.cookie.split('; ').map(c => {
      const [key, ...rest] = c.split('=');
      return [key, rest.join('=')];
    })
  );
}

function getToken(): string | null {
  const stored = localStorage.getItem('oscar_auth');
  if (stored) {
    const { token } = JSON.parse(stored);
    if (token) return token;
  }
  const cookies = parseCookies();
  return cookies['oscar_token'] || null;
}