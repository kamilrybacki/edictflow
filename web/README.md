# Edictflow Web UI

The web frontend for Edictflow - a centralized CLAUDE.md configuration management system.

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Next.js | 16.1.6 | React framework with App Router |
| React | 19.2.3 | UI library |
| Tailwind CSS | 4 | Utility-first styling |
| React Flow | 11.11.4 | Graph visualization |
| Playwright | 1.58.1 | E2E testing |
| Jest | 29.7.0 | Unit testing |
| TypeScript | 5 | Type safety |

## Getting Started

### Prerequisites

- Node.js 20+
- npm or yarn

### Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build

```bash
# Production build
npm run build

# Start production server
npm start
```

## Project Structure

```
web/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── admin/              # Admin pages
│   │   │   └── teams/          # Team management
│   │   ├── api/                # API routes
│   │   ├── approvals/          # Approvals page
│   │   ├── changes/            # Changes audit log
│   │   ├── graph/              # Graph view page
│   │   ├── login/              # Authentication
│   │   ├── register/           # User registration
│   │   ├── settings/           # User settings
│   │   ├── layout.tsx          # Root layout
│   │   └── page.tsx            # Dashboard
│   ├── components/
│   │   ├── dashboard/          # Dashboard components
│   │   ├── graph/              # Graph view components
│   │   ├── RuleEditor/         # Rule editing components
│   │   ├── ui/                 # Shared UI components
│   │   ├── CommandPalette.tsx  # Ctrl+K command palette
│   │   └── ...
│   ├── contexts/               # React contexts
│   │   └── AuthContext.tsx     # Authentication state
│   ├── domain/                 # Domain types
│   │   └── rule.ts             # Rule, Role types
│   ├── lib/
│   │   ├── api/                # API client functions
│   │   └── layerConfig.ts      # Configuration
│   └── __tests__/              # Unit tests
├── e2e/                        # Playwright E2E tests
│   ├── home.spec.ts
│   ├── graph-view.spec.ts
│   ├── approvals.spec.ts
│   ├── changes.spec.ts
│   ├── admin-features.spec.ts
│   └── user-flows.spec.ts
└── public/                     # Static assets
```

## Key Features

### Dashboard

The main dashboard (`/`) provides:

- System overview with stats
- Active rules summary
- Connected agents count
- Recent activity

### Graph View

Interactive organization visualization (`/graph`):

- Teams, users, and rules as nodes
- Hierarchical layout with dagre
- Filter by team or rule status
- Zoom, pan, and fit controls
- Node selection and highlighting

### Command Palette

Quick navigation with `Ctrl+K` / `Cmd+K`:

- Search pages, teams, and rules
- Keyboard-driven navigation
- Recent destinations

### Rule Editor

Create and edit CLAUDE.md rules:

- Content editing
- Enforcement mode selection
- Target teams/users configuration
- Trigger pattern configuration

## Available Scripts

| Script | Description |
|--------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Production build |
| `npm start` | Start production server |
| `npm run lint` | Run ESLint |
| `npm test` | Run Jest unit tests |
| `npm run test:watch` | Run tests in watch mode |
| `npm run test:coverage` | Run tests with coverage |
| `npm run test:e2e` | Run Playwright E2E tests |
| `npm run test:e2e:ui` | Run E2E tests with UI |
| `npm run test:e2e:headed` | Run E2E tests headed |

## Testing

### Unit Tests

```bash
npm test
```

Tests are located in `src/__tests__/`.

### E2E Tests

```bash
npm run test:e2e
```

Tests are located in `e2e/`. Uses `alex.rivera@test.local` as the test user.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend API URL |

## Development Notes

### Adding a New Page

1. Create page in `src/app/[route]/page.tsx`
2. Add to command palette in `CommandPalette.tsx`
3. Add navigation link in layout if needed
4. Add E2E tests in `e2e/`

### Adding a New Component

1. Create in `src/components/`
2. Add unit tests in `src/__tests__/components/`
3. Export from appropriate index if shared

### Graph View Development

The graph uses React Flow with dagre for layout:

- `src/components/graph/GraphView.tsx` - Main graph component
- `src/components/graph/GraphControls.tsx` - Filter controls
- `src/lib/api/graph.ts` - API client

## Related Documentation

- [Web UI Guide](../docs/web-ui/index.md) - User documentation
- [Graph View Feature](../docs/features/graph.md) - Graph feature docs
- [Testing Guide](../docs/development/testing.md) - Testing strategy
