# oscar Frontend

A modern CRM frontend built with Astro and React, following the Islands Architecture pattern.

## Tech Stack

- **Astro 6** — Meta-framework for SSR, routing, and layouts
- **React 10** — Interactive islands for complex UI components
- **TypeScript** — Full type safety across the codebase
- **Tailwind CSS 4** — Utility-first styling
- **Nanostores** — Shared state management across React islands and Astro
- **Lucide React** — Icon library

## Project Structure

```
src/
├── components/
│   ├── ui/           # Reusable UI primitives (Button, Modal, Input, Badge)
│   └── layout/       # Layout-specific components (AppLayout, NavBar)
│
├── islands/          # React components with client-side interactivity
│   ├── deals/        # Deal-specific islands (DealKanban, DealForm, DealModal)
│   ├── contacts/    # Contact-specific islands (ContactsTable, ContactForm)
│   ├── activities/  # Activity-specific islands (ActivityTimeline, ActivityForm)
│   ├── companies/   # Company-specific islands
│   ├── teams/       # Team management islands
│   └── shared/      # Shared interactive components
│
├── layouts/          # Astro page layouts
│   └── AppLayout.astro  # Authenticated app shell with sidebar/header
│
├── lib/              # Core utilities
│   ├── api.ts        # API client with fetch wrapper
│   ├── stores.ts     # Nanostores global state
│   └── utils.ts      # Utility functions
│
├── pages/            # Astro file-based routing
│   ├── index.astro           # Root redirect
│   ├── login.astro           # Login page
│   ├── register.astro        # Registration page
│   ├── dashboard.astro      # Dashboard with KPIs, charts, activity
│   ├── deals/
│   │   └── index.astro      # Deal pipeline with kanban board
│   ├── contacts/
│   │   ├── index.astro      # Contact list with table
│   │   └���─ [id].astro       # Contact detail (stub)
│   ├── companies/
│   │   └── index.astro      # Company grid with cards
│   ├── activities/
│   │   └── index.astro      # Activity timeline
│   ├── teams/
│   │   └── index.astro      # Team management
│   ├── automations/
│   │   └── index.astro      # Automation builder (static UI)
│   ├── settings/
│   │   ├── index.astro      # Main settings
│   │   ├── users.astro      # User management
│   │   ├── security.astro  # Security settings
│   │   ├── pipelines.astro  # Pipeline config
│   │   └── profile.astro    # Profile settings
│   ├── invite/[token].astro       # Invitation page
│   ├── verify-email/[token].astro # Email verification
│   └── resend-verification.astro  # Resend verification
│
├── types/            # TypeScript type definitions
│   ├── deal.ts
│   ├── person.ts
│   ├── company.ts
│   ├── activity.ts
│   ├── team.ts
│   ├── user.ts
│   ├── auth.ts
│   └── api.ts
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

- **Button** — Variants: primary, secondary, danger, ghost, outline. Sizes: sm, md, lg
- **Modal** — Accessible dialog with backdrop, escape to close
- **Input / Textarea / Select** — Consistent styling with labels and errors
- **Badge** — Status indicators with color variants
- **Avatar** — User avatars with fallback initials
- **Card** — Container with header, body, footer sections

### Layout

**AppLayout** — Authenticated app shell with:
- Sidebar navigation with module links
- Header with notification bell and user menu
- Main content area with padding
- Mobile-responsive sidebar (collapsible)

**TopNavBar** — Header with:
- Notification bell with unread count
- User avatar and dropdown menu
- Search trigger (command palette)

**SideNavBar** — Navigation with:
- Dashboard link
- CRM section (Contacts, Companies, Deals)
- Activity section (Activities, Automations)
- Team section (Teams)
- Settings section

## Pages

Each page is an Astro component with optional React islands for interactivity:

| Route | Status | Description |
|-------|--------|-------------|
| `/` | Done | Root — redirects to dashboard or login |
| `/login` | Done | Login form with email/password |
| `/register` | Done | Registration form with tenant setup |
| `/dashboard` | Done | Overview with stats, charts, recent activity, task modal |
| `/deals` | Done | Pipeline view with kanban board, drag-drop |
| `/contacts` | Done | Contact list with table, filters, bulk actions, modals |
| `/contacts/[id]` | Stub | Contact detail page (needs real data) |
| `/companies` | Done | Company cards grid with create/edit modal |
| `/activities` | Done | Activity timeline with filters, log activity modal |
| `/teams` | Done | Team member list, invite modal, edit modal |
| `/automations` | Static | Automation builder (static UI mockup) |
| `/settings` | Done | Branding, colors, fonts, logo uploads |
| `/settings/users` | Done | User management |
| `/settings/security` | Done | Security settings |
| `/settings/pipelines` | Stub | Pipeline configuration |
| `/settings/profile` | Done | Profile settings |
| `/invite/[token]` | Done | Invitation acceptance |
| `/verify-email/[token]` | Done | Email verification |
| `/resend-verification` | Done | Resend verification email |

## Development

```bash
# Install dependencies
npm install

# Start all services (uses launch.sh)
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

## Design Principles

1. **Server-first** — SSR for initial load, hydrate only interactive parts
2. **No component library** — Custom Tailwind components for consistency
3. **Feature-based** — Group by domain, not by file type
4. **Type-safe** — TypeScript everywhere, shared types with backend
5. **Progressive enhancement** — Pages work without JavaScript, islands add interactivity

## Build

```bash
npm run build    # Production build
npm run preview  # Preview production build
```

## Roadmap

### Implemented

- [x] Core CRM pages (contacts, companies, deals, activities)
- [x] Dashboard with KPIs and charts
- [x] Kanban board for deals with drag-drop
- [x] Activity timeline
- [x] Team management with invitations
- [x] Settings with white-label branding
- [x] Command palette / quick search
- [x] Notifications panel

### Pending Implementation

- [ ] **Contact/Company/Deal Detail Views** — Real data display with tabs
- [ ] **Custom Fields UI** — Dynamic field rendering
- [ ] **Reports & Analytics Page** — Dedicated reports beyond dashboard
- [ ] **Workflow Builder** — Visual automation editor (static UI exists)
- [ ] **Product Catalog Page** — Product management UI
- [ ] **Calendar View** — Monthly/weekly activity view
- [ ] **Global Search Page** — Full-text search across all entities
- [ ] **CSV Import Wizard** — Import contacts, companies, deals
- [ ] **Bulk Edit Operations** — Multi-record editing
- [ ] **Keyboard Shortcuts** — Command palette expansion
- [ ] **Mobile Responsive** — Fully responsive layouts