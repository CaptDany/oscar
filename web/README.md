# oscar Frontend

A modern CRM frontend built with Astro and React, following the Islands Architecture pattern.

## Tech Stack

- **Astro** — Meta-framework for SSR, routing, and layouts
- **React** — Interactive islands for complex UI components
- **TypeScript** — Full type safety across the codebase
- **Tailwind CSS** — Utility-first styling
- **Nanostores** — Shared state management across React islands and Astro
- **Lucide React** — Icon library

## Project Structure

```
src/
├── components/
│   ├── ui/           # Reusable UI primitives (Button, Modal, Input, Badge)
│   └── layout/        # Layout-specific components
│
├── islands/          # React components with client-side interactivity
│   ├── deals/        # Deal-specific islands
│   ├── contacts/     # Contact-specific islands
│   └── shared/       # Shared interactive components
│
├── layouts/          # Astro page layouts
│   └── AppLayout.astro  # Authenticated app shell
│
├── lib/              # Core utilities
│   ├── api.ts        # API client
│   ├── auth.ts       # Auth store (legacy)
│   └── stores.ts     # Nanostores global state
│
├── pages/            # Astro file-based routing
│   ├── index.astro   # Root redirect
│   ├── login.astro   # Login page
│   ├── register.astro # Registration page
│   ├── dashboard.astro # Dashboard
│   ├── deals/        # Deals module
│   ├── contacts/     # Contacts module
│   ├── companies/    # Companies module
│   └── settings/    # Settings module
│
├── types/            # TypeScript type definitions
│   ├── deal.ts
│   └── person.ts
│
├── middleware.ts     # Astro middleware (auth check)
└── env.d.ts          # TypeScript declarations
```

## Architecture

### Islands Architecture

Astro renders pages server-side by default, providing excellent initial load performance. Interactive components (forms, modals, kanban boards) are React islands that hydrate only when needed.

**Hydration strategies:**
- `client:load` — Hydrate immediately on page load
- `client:visible` — Hydrate when component enters viewport
- `client:idle` — Hydrate when browser is idle

### State Management

Nanostores for cross-boundary state. Since Astro + React islands can't share React context, we use Nanostores atoms and maps:

```typescript
// lib/stores.ts
import { atom, map } from 'nanostores';

export const $auth = map<AuthState>({...});

// In React island
import { useStore } from '@nanostores/react';
import { $auth } from '../../lib/stores';

export function DealForm() {
  const auth = useStore($auth);
  // ...
}
```

### API Client

All API calls go through the shared client in `lib/api.ts`. Uses relative URLs (`/api/v1/...`) — the Astro dev server proxies to Go backend via Vite proxy.

## Components

### UI Primitives

Built from scratch with Tailwind, no component library:

- **Button** — Variants: primary, secondary, danger, ghost. Sizes: sm, md, lg
- **Modal** — Accessible dialog with backdrop, escape to close
- **Input / Textarea / Select** — Consistent styling with labels and errors
- **Badge** — Status indicators with color variants

### Layout

**AppLayout** — Authenticated app shell with:
- Sidebar navigation
- Header with notification bell
- Main content area

## Pages

Each page is an Astro component with optional React islands for interactivity:

| Route | Description |
|-------|-------------|
| `/` | Root — redirects to dashboard or login |
| `/login` | Login form |
| `/register` | Registration form |
| `/dashboard` | Overview with stats and recent activity |
| `/deals` | Pipeline view with kanban board |
| `/contacts` | Contact list with search |
| `/companies` | Company cards grid |
| `/settings` | User and organization settings |

## Development

```bash
# Start all services
cd .. && ./launch.sh start

# The frontend runs at http://localhost:4321
# API requests are proxied to Go backend at localhost:8080
```

### Adding a New Page

1. Create `src/pages/{module}/index.astro`
2. Use `AppLayout` for authenticated pages
3. Add navigation link in sidebar
4. For interactive features, create React components in `islands/`

```astro
---
import AppLayout from '../../layouts/AppLayout.astro';
import { DealForm } from '../../islands/deals/DealForm';
---

<AppLayout title="Deals">
  <DealForm client:load />
</AppLayout>
```

### Adding a New Component

1. UI primitives go in `src/components/ui/`
2. Page-specific components go in `src/islands/{module}/`
3. Types go in `src/types/`

## Build

```bash
npm run build    # Production build
npm run preview  # Preview production build
```

## Design Principles

1. **Server-first** — SSR for initial load, hydrate only interactive parts
2. **No component library** — Custom Tailwind components for consistency
3. **Feature-based** — Group by domain, not by file type
4. **Type-safe** — TypeScript everywhere, shared types between frontend and potential backend
5. **Progressive enhancement** — Pages work without JavaScript, islands add interactivity
