# Graph View Design

**Date**: 2026-02-07
**Status**: Draft
**Feature**: Visualize relationships between users, teams, and rules

## Overview

A standalone graph visualization page at `/graph` that shows how rules target users and teams. Users can filter, pan/zoom, and click nodes to explore connections.

## Use Case

Primary: "Which rules affect which users and teams?"

Helps administrators understand targeting relationships at a glance without navigating between multiple detail pages.

## Data Model

### Node Types

| Type | Shape | Color | Display |
|------|-------|-------|---------|
| Team | Rounded rectangle | Blue (#3B82F6) | Name + member count |
| User | Circle | Green (#22C55E) | Avatar/initials + name |
| Rule | Hexagon | Amber (#F59E0B) | Name + status badge |

### Edge Types

| Relationship | Source | Target | Style |
|--------------|--------|--------|-------|
| Belongs to | User | Team | Solid gray |
| Targets | Rule | Team/User | Dashed blue with arrow |
| Owned by | Rule | Team | Dotted gray |

### Rule Status Indicators

- **Draft**: Dashed border
- **Pending**: Pulsing/animated border
- **Approved**: Solid border + checkmark badge
- **Rejected**: Red border + X badge

## Layout

**Hierarchical arrangement:**
- Top row: Teams
- Middle row: Users
- Bottom row: Rules

Edges flow downward from rules to their targets, upward from users to teams.

## Page Structure

```
┌─────────────────────────────────────────────────────────┐
│  Organization Graph    [Team ▼] [Status ▼] [Search...] │
├─────────────────────────────────────────────────────────┤
│                                                         │
│                    Graph Canvas                         │
│                                                         │
│                                              ┌────────┐ │
│                                              │Minimap │ │
│                                              └────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Controls

**Filter bar (top):**
- Team dropdown: Filter to selected team(s) and connections
- Rule status toggle: Show/hide by status (draft, pending, approved, rejected)
- Search box: Find and focus on entity by name

**Viewport (React Flow built-ins):**
- Zoom in/out buttons
- Fit-to-view button
- Pan via drag

## Interactions

### Click Node

1. Highlight all connected edges (dim unconnected to 30% opacity)
2. Show popover with:
   - Name and type icon
   - Key info:
     - Team: member count
     - User: email, team name
     - Rule: status, enforcement mode
   - "View details" link to entity page

### Click Empty Canvas

Clear selection, restore full opacity.

### Hover Node

Subtle glow effect, cursor → pointer.

### Hover Edge

Thicken edge, show tooltip with relationship type.

## API

### GET /api/graph

Single endpoint returns all graph data:

```json
{
  "teams": [
    { "id": "uuid", "name": "Platform", "memberCount": 5 }
  ],
  "users": [
    { "id": "uuid", "name": "Alice", "email": "alice@example.com", "teamId": "uuid" }
  ],
  "rules": [
    {
      "id": "uuid",
      "name": "No secrets in code",
      "status": "approved",
      "enforcementMode": "block",
      "teamId": "uuid",
      "targetTeams": ["uuid"],
      "targetUsers": ["uuid"]
    }
  ]
}
```

**Permissions:** Returns only data user has permission to view (respects existing RBAC).

## File Structure

### Frontend (web/)

```
web/src/app/graph/page.tsx              # Page component
web/src/components/graph/
  ├── GraphView.tsx                     # Main React Flow canvas
  ├── GraphControls.tsx                 # Filter bar
  ├── nodes/
  │   ├── TeamNode.tsx                  # Custom team node
  │   ├── UserNode.tsx                  # Custom user node
  │   └── RuleNode.tsx                  # Custom rule node
  └── GraphPopover.tsx                  # Click popover
```

### Backend (server/)

```
server/entrypoints/api/handlers/graph.go   # Handler
```

### Dependencies

- Add `reactflow` package

## Out of Scope

- Editing from graph (view-only)
- Real-time updates (refresh to see changes)
- Saving graph layout positions
- Export to image/PDF

## Technical Notes

### Library Choice: React Flow

Selected over D3.js and Cytoscape.js because:
- Native React integration with hooks-based API
- Built-in TypeScript support
- Includes pan/zoom, minimap, controls out of box
- Handles moderate graph size (hundreds of nodes) well

### Layout Algorithm

Use React Flow's `dagre` layout plugin for hierarchical positioning:
- Direction: top-to-bottom
- Node separation tuned for readability
- Recalculates on filter changes

### Dark Mode

Uses existing Tailwind dark mode classes. Nodes use darker backgrounds with light text in dark mode.
