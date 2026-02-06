# Command Palette Search Design

## Overview

Add command palette functionality (⌘K) to the dashboard search bar. Currently the search input is decorative - this makes it functional with a modal overlay for searching rules, teams, and quick actions.

## Core Behavior

**Trigger:**
- Click on search input in header
- Press `⌘K` (Mac) / `Ctrl+K` (Windows/Linux)
- Opens modal overlay with search input at top

**Categories:**
1. **Rules** - search by name, description, status, target layer
2. **Teams** - search by name
3. **Quick Actions** - Create Rule, Create Team, View Approvals, View Agents, Logout

**Interaction:**
- Type to filter results across all categories
- Arrow keys to navigate, Enter to select
- Escape or click outside to close
- Results grouped by category with headers

## Suggestion Behavior

- **0-2 characters:** Show quick actions only (no search)
- **3+ characters:** Trigger search, show filtered results by category

**Matching:**
- Case-insensitive substring matching
- Match: rule name, rule description, team name
- Actions shown if name contains query
- 150ms debounce on typing

**Ordering:**
1. Exact prefix matches first
2. Substring matches second
3. Recent/relevant items prioritized

**Limits:**
- Max 5 rules, 3 teams, all actions displayed

## Result Display

| Category | Icon | Primary | Secondary | On Select |
|----------|------|---------|-----------|-----------|
| Rules | Layer icon | Rule name | Status + layer | Select rule, show details |
| Teams | Users icon | Team name | Member count | Switch to team |
| Actions | Action icon | Action name | Shortcut hint | Execute action |

## Visual Design

- Semi-transparent backdrop (bg-black/50)
- Centered modal, max-width 500px
- Search input with Search icon at top
- Scrollable results area
- Keyboard shortcut hints on actions
- Highlight matched text in bold

## Component Structure

**New file:**
```
web/src/components/CommandPalette.tsx
```

**Props:**
```typescript
interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  rules: Rule[];
  teams: TeamData[];
  onSelectRule: (rule: Rule) => void;
  onSelectTeam: (team: TeamData | null) => void;
  onCreateRule: () => void;
  onCreateTeam: () => void;
  onViewApprovals: () => void;
  onViewAgents: () => void;
  onLogout: () => void;
}
```

**Changes to existing files:**

`DashboardLayout.tsx`:
- Add `isOpen` state for command palette
- Add global keyboard listener for ⌘K/Ctrl+K
- Make search input click open palette instead of typing
- Accept new props for rules, teams, and action handlers
- Render CommandPalette component

`page.tsx`:
- Pass `rules` and `teams` to DashboardLayout
- Pass action handlers to DashboardLayout

## Implementation Steps

1. Create `CommandPalette.tsx` component
   - Modal structure with backdrop
   - Search input with debounced state
   - Results filtering logic
   - Keyboard navigation
   - Result rendering by category

2. Update `DashboardLayout.tsx`
   - Add CommandPalette import and state
   - Add keyboard shortcut listener (useEffect)
   - Modify search input to open palette on click
   - Add new props to interface
   - Render CommandPalette

3. Update `page.tsx`
   - Add router.push handler for approvals
   - Pass all required props to DashboardLayout

4. Test interactions
   - ⌘K opens palette
   - Typing filters results after 3 chars
   - Arrow keys navigate
   - Enter selects
   - Escape closes
   - Actions execute correctly
