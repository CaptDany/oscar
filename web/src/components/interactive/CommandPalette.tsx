import { useState, useEffect, useRef } from 'preact/hooks';

interface RecentItem {
  type: string;
  name: string;
  href: string;
  time: string;
}

interface SearchResult {
  type: 'Deal' | 'Contact' | 'Company';
  name: string;
  href: string;
  value?: number;
}

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [_recent, setRecent] = useState<RecentItem[]>([]);
  const [results, setResults] = useState<SearchResult[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const handleKeydown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setOpen(true);
      }
      if (e.key === 'Escape') {
        setOpen(false);
      }
    };

    document.addEventListener('keydown', handleKeydown);

    const stored = localStorage.getItem('oscar_recent');
    if (stored) {
      try {
        setRecent(JSON.parse(stored));
      } catch (e) {}
    }

    return () => document.removeEventListener('keydown', handleKeydown);
  }, []);

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus();
    }
  }, [open]);

  useEffect(() => {
    if (query.length < 2) {
      setResults([]);
      return;
    }

    const timeout = setTimeout(async () => {
      const token = getToken();
      if (!token) return;

      try {
        const res = await fetch(`/api/v1/search?q=${encodeURIComponent(query)}&limit=10`, {
          headers: { Authorization: `Bearer ${token}` }
        });

        if (!res.ok) return;

        const data = await res.json();

        const searchResults: SearchResult[] = (data.data || []).map((r: any) => ({
          type: (r.entity_type === 'person' ? 'Contact' : r.entity_type === 'deal' ? 'Deal' : 'Company') as 'Deal' | 'Contact' | 'Company',
          name: r.title,
          href: r.entity_type === 'person' ? `/contacts/${r.id}` : r.entity_type === 'company' ? `/companies/${r.id}` : `/deals/${r.id}`,
          value: r.entity_type === 'deal' ? undefined : undefined,
        }));

        setResults(searchResults.slice(0, 10));
      } catch (err) {
        console.error('Search failed:', err);
      }
    }, 300);

    return () => clearTimeout(timeout);
  }, [query]);

  const executeCommand = (command: string) => {
    setOpen(false);
    const routes: Record<string, string> = {
      'new-deal': '/deals',
      'new-contact': '/contacts',
      'goto-dashboard': '/dashboard',
      'goto-contacts': '/contacts',
      'goto-deals': '/deals',
      'goto-settings': '/settings',
    };
    if (routes[command]) {
      window.location.href = routes[command];
    }
  };

  if (!open) return null;

  return (
    <div class="fixed inset-0 z-[200] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-background/60 backdrop-blur-md"
        onClick={() => setOpen(false)}
      />
      <div class="relative w-full max-w-2xl bg-surface-container border border-outline-variant/20 rounded-xl shadow-2xl shadow-black/80 overflow-hidden flex flex-col">
        <div class="px-6 py-5 border-b border-outline-variant/10 flex items-center gap-4">
          <span class="material-symbols-outlined text-primary text-2xl" style="font-variation-settings: 'FILL' 1;">bolt</span>
          <input
            ref={inputRef}
            class="flex-1 bg-transparent border-none focus:ring-0 text-xl font-body placeholder:text-gray-600 text-on-background"
            placeholder="Type a command or search records..."
            type="text"
            value={query}
            onInput={(e) => setQuery((e.target as HTMLInputElement).value)}
          />
          <div class="flex items-center gap-1.5 px-2 py-1 bg-surface-container-highest border border-outline-variant/20 rounded text-[10px] font-mono text-gray-400">
            <span>ESC</span>
          </div>
        </div>

        <div class="flex-1 max-h-[614px] overflow-y-auto custom-scrollbar p-2">
          {results.length > 0 ? (
            <div class="mb-4">
              <div class="px-4 py-2 text-[10px] font-mono text-gray-500 uppercase tracking-[0.2em]">Search Results</div>
              <div class="space-y-0.5">
                {results.map((item, i) => (
                  <button
                    key={i}
                    onClick={() => window.location.href = item.href}
                    class="w-full flex items-center gap-4 px-4 py-3 rounded-lg hover:bg-white/5 group transition-all text-left"
                  >
                    <div class="w-8 h-8 rounded bg-primary/10 flex items-center justify-center">
                      <span class="material-symbols-outlined text-primary text-lg">
                        {item.type === 'Deal' ? 'attach_money' : item.type === 'Contact' ? 'person' : 'domain'}
                      </span>
                    </div>
                    <div class="flex-1">
                      <div class="text-sm font-medium text-white">{item.name}</div>
                      <div class="text-[10px] font-mono text-gray-500">
                        {item.type}{item.value ? ` • $${item.value.toLocaleString()}` : ''}
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div class="mb-4">
              <div class="px-4 py-2 text-[10px] font-mono text-gray-500 uppercase tracking-[0.2em]">Quick Actions</div>
              <div class="space-y-0.5">
                {[
                  { cmd: 'new-deal', icon: 'add_circle', label: 'Create New Deal', shortcut: 'CMD + N' },
                  { cmd: 'new-contact', icon: 'person_add', label: 'Add Contact', shortcut: 'CMD + Shift + N' },
                  { cmd: 'goto-dashboard', icon: 'dashboard', label: 'Go to Dashboard', shortcut: 'G' },
                  { cmd: 'goto-contacts', icon: 'group', label: 'Go to Contacts', shortcut: 'C' },
                  { cmd: 'goto-deals', icon: 'account_tree', label: 'Go to Deals', shortcut: 'D' },
                  { cmd: 'goto-settings', icon: 'settings', label: 'Go to Settings', shortcut: ',' },
                ].map((action) => (
                  <button
                    key={action.cmd}
                    onClick={() => executeCommand(action.cmd)}
                    class="w-full flex items-center gap-4 px-4 py-3 rounded-lg hover:bg-white/5 group transition-all text-left"
                  >
                    <div class="w-8 h-8 rounded bg-primary/10 flex items-center justify-center">
                      <span class="material-symbols-outlined text-primary text-lg">{action.icon}</span>
                    </div>
                    <div class="flex-1">
                      <div class="text-sm font-medium text-white">{action.label}</div>
                      <div class="text-[10px] font-mono text-primary/60">{action.shortcut}</div>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>

        <div class="px-6 py-3 bg-surface-container-low border-t border-outline-variant/10 flex justify-between items-center">
          <div class="flex items-center gap-4 text-[10px] font-mono text-gray-500">
            <span class="flex items-center gap-1.5">
              <span class="material-symbols-outlined text-[14px]">keyboard_arrow_up</span>
              <span class="material-symbols-outlined text-[14px]">keyboard_arrow_down</span>
              <span>Navigate</span>
            </span>
            <span class="flex items-center gap-1.5">
              <span class="material-symbols-outlined text-[14px]">keyboard_return</span>
              <span>Open</span>
            </span>
          </div>
          <div class="text-[10px] font-mono text-primary/40 uppercase tracking-widest">OSCAR v0.1.0</div>
        </div>
      </div>
    </div>
  );
}

function getToken(): string | null {
  const stored = localStorage.getItem('oscar_auth');
  if (stored) {
    const { token } = JSON.parse(stored);
    if (token) return token;
  }
  const cookies = Object.fromEntries(
    document.cookie.split('; ').map(c => c.split('='))
  );
  return cookies['oscar_token'] || null;
}