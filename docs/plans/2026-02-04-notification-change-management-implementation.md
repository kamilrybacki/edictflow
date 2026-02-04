# Implementation Plan: Notification & Change Management System

Based on: [Notification & Change Management System Design](./2026-02-04-notification-change-management-design.md)

## Requirements Restatement

Based on the design document, this system provides:

1. **Change Detection & Enforcement** - Agent monitors CLAUDE.md files and enforces policies (block/temporary/warning modes)
2. **Change Request Lifecycle** - Users' local modifications create pending requests that admins approve/reject
3. **Exception System** - Users can appeal rejected changes with justifications
4. **Bidirectional Notifications** - Desktop notifications to users, web/email/webhook alerts to admins
5. **Notification Channels** - Configurable per-team delivery (email, webhooks, in-app)

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| **WebSocket reliability** - Messages lost during disconnection | HIGH | Queue pending messages, replay on reconnect, SQLite local cache |
| **Auto-revert race conditions** - User edits during pending state | MEDIUM | Lock file during revert, update existing request on re-edit |
| **Email/webhook failures** - External services unavailable | MEDIUM | Retry with exponential backoff, don't block main workflow |
| **Large diffs** - Performance with huge file changes | LOW | Truncate diff_content at reasonable limit (64KB) |
| **Notification spam** - Too many alerts | LOW | Rate limiting, digest mode for webhooks |

---

## Implementation Phases

### Phase 1: Database Schema & Migrations ✅ COMPLETED

**Objective:** Add the four new tables and extend rules table

**Files to create:**
- `server/migrations/000015_add_change_requests.up.sql`
- `server/migrations/000015_add_change_requests.down.sql`
- `server/migrations/000016_add_exception_requests.up.sql`
- `server/migrations/000016_add_exception_requests.down.sql`
- `server/migrations/000017_add_notifications.up.sql`
- `server/migrations/000017_add_notifications.down.sql`
- `server/migrations/000018_add_notification_channels.up.sql`
- `server/migrations/000018_add_notification_channels.down.sql`
- `server/migrations/000019_add_rule_enforcement.up.sql`
- `server/migrations/000019_add_rule_enforcement.down.sql`

**Schema (from design doc):**
```sql
-- Change requests from local modifications
CREATE TABLE change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES rules(id),
    agent_id UUID NOT NULL REFERENCES agents(id),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID NOT NULL REFERENCES teams(id),
    file_path TEXT NOT NULL,
    original_hash TEXT NOT NULL,
    modified_hash TEXT NOT NULL,
    diff_content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    enforcement_mode TEXT NOT NULL,
    timeout_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

-- Exception requests for rejected changes
CREATE TABLE exception_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id UUID NOT NULL REFERENCES change_requests(id),
    user_id UUID NOT NULL REFERENCES users(id),
    justification TEXT NOT NULL,
    exception_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id)
);

-- Notifications for users
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    team_id UUID REFERENCES teams(id),
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Notification channels (email, webhooks)
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    channel_type TEXT NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add to existing rules table
ALTER TABLE rules
ADD COLUMN enforcement_mode TEXT NOT NULL DEFAULT 'block',
ADD COLUMN temporary_timeout_hours INTEGER NOT NULL DEFAULT 24;
```

**Indexes needed:**
- `change_requests`: `(team_id, status)`, `(rule_id)`, `(agent_id)`, `(timeout_at)` for worker queries
- `exception_requests`: `(change_request_id)`, `(status)`
- `notifications`: `(user_id, read_at)`, `(team_id)`
- `notification_channels`: `(team_id, enabled)`

---

### Phase 2: Domain Models ✅ COMPLETED

**Objective:** Define Go domain entities

**Files to create:**
- `server/domain/change_request.go`
- `server/domain/exception_request.go`
- `server/domain/notification.go`
- `server/domain/notification_channel.go`

**Extend:**
- `server/domain/rule.go` - Add `EnforcementMode` and `TemporaryTimeoutHours` fields

**Types/Constants:**
```go
// ChangeRequestStatus
const (
    ChangeRequestStatusPending          = "pending"
    ChangeRequestStatusApproved         = "approved"
    ChangeRequestStatusRejected         = "rejected"
    ChangeRequestStatusAutoReverted     = "auto_reverted"
    ChangeRequestStatusExceptionGranted = "exception_granted"
)

// EnforcementMode
const (
    EnforcementModeBlock     = "block"
    EnforcementModeTemporary = "temporary"
    EnforcementModeWarning   = "warning"
)

// ExceptionType
const (
    ExceptionTypeTimeLimited = "time_limited"
    ExceptionTypePermanent   = "permanent"
)

// NotificationType
const (
    NotificationTypeChangeDetected     = "change_detected"
    NotificationTypeApprovalRequired   = "approval_required"
    NotificationTypeChangeApproved     = "change_approved"
    NotificationTypeChangeRejected     = "change_rejected"
    NotificationTypeChangeAutoReverted = "change_auto_reverted"
    NotificationTypeExceptionGranted   = "exception_granted"
    NotificationTypeExceptionDenied    = "exception_denied"
)
```

---

### Phase 3: Repository Layer ✅ COMPLETED

**Objective:** Database adapters for new entities

**Files to create:**
- `server/adapters/postgres/change_request_db.go`
- `server/adapters/postgres/exception_request_db.go`
- `server/adapters/postgres/notification_db.go`
- `server/adapters/postgres/notification_channel_db.go`

**Repository interfaces to define (in service layer):**
```go
type ChangeRequestRepository interface {
    Create(ctx context.Context, cr *ChangeRequest) error
    GetByID(ctx context.Context, id uuid.UUID) (*ChangeRequest, error)
    ListByTeam(ctx context.Context, teamID uuid.UUID, filter ChangeRequestFilter) ([]ChangeRequest, error)
    Update(ctx context.Context, cr *ChangeRequest) error
    FindExpiredTemporary(ctx context.Context, before time.Time) ([]ChangeRequest, error)
    FindByAgentAndFile(ctx context.Context, agentID uuid.UUID, filePath string) (*ChangeRequest, error)
}

type ExceptionRequestRepository interface {
    Create(ctx context.Context, er *ExceptionRequest) error
    GetByID(ctx context.Context, id uuid.UUID) (*ExceptionRequest, error)
    ListByTeam(ctx context.Context, teamID uuid.UUID, filter ExceptionRequestFilter) ([]ExceptionRequest, error)
    Update(ctx context.Context, er *ExceptionRequest) error
    FindActiveByUserRuleFile(ctx context.Context, userID, ruleID uuid.UUID, filePath string) (*ExceptionRequest, error)
}

type NotificationRepository interface {
    Create(ctx context.Context, n *Notification) error
    CreateBulk(ctx context.Context, notifications []Notification) error
    GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
    ListByUser(ctx context.Context, userID uuid.UUID, filter NotificationFilter) ([]Notification, error)
    MarkRead(ctx context.Context, id uuid.UUID) error
    MarkAllRead(ctx context.Context, userID uuid.UUID) error
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
}

type NotificationChannelRepository interface {
    Create(ctx context.Context, nc *NotificationChannel) error
    GetByID(ctx context.Context, id uuid.UUID) (*NotificationChannel, error)
    ListByTeam(ctx context.Context, teamID uuid.UUID) ([]NotificationChannel, error)
    Update(ctx context.Context, nc *NotificationChannel) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListEnabledByTeam(ctx context.Context, teamID uuid.UUID) ([]NotificationChannel, error)
}
```

---

### Phase 4: Core Services ✅ COMPLETED

**Objective:** Business logic for change/exception/notification lifecycle

**Directory structure:**
```
server/services/
├── changes/
│   ├── service.go       # ChangeRequest lifecycle
│   └── repository.go    # Interface definitions
├── exceptions/
│   ├── service.go       # ExceptionRequest lifecycle
│   └── repository.go
└── notifications/
    ├── service.go       # Create, query, mark-read
    ├── repository.go
    └── dispatcher.go    # Background delivery (email/webhook)
```

**ChangeService methods:**
```go
type ChangeService interface {
    // Agent reports a change
    CreateFromAgent(ctx context.Context, payload AgentChangePayload) (*ChangeRequest, error)

    // Admin actions
    Approve(ctx context.Context, id, approverUserID uuid.UUID) error
    Reject(ctx context.Context, id, approverUserID uuid.UUID) error

    // Worker calls for auto-revert
    HandleExpiredTemporary(ctx context.Context) ([]ChangeRequest, error)

    // User re-edits pending file
    UpdateFromAgent(ctx context.Context, id uuid.UUID, newDiff string, newHash string) error

    // Queries
    GetByID(ctx context.Context, id uuid.UUID) (*ChangeRequest, error)
    ListByTeam(ctx context.Context, teamID uuid.UUID, filter ChangeRequestFilter) ([]ChangeRequest, error)
}
```

**ExceptionService methods:**
```go
type ExceptionService interface {
    // User creates exception request
    Create(ctx context.Context, req CreateExceptionRequest) (*ExceptionRequest, error)

    // Admin actions
    Approve(ctx context.Context, id, approverUserID uuid.UUID, expiresAt *time.Time) error
    Deny(ctx context.Context, id, approverUserID uuid.UUID) error

    // Check if exception exists
    HasActiveException(ctx context.Context, userID, ruleID uuid.UUID, filePath string) (bool, error)

    // Queries
    GetByID(ctx context.Context, id uuid.UUID) (*ExceptionRequest, error)
    ListByTeam(ctx context.Context, teamID uuid.UUID, filter ExceptionRequestFilter) ([]ExceptionRequest, error)
}
```

**NotificationService methods:**
```go
type NotificationService interface {
    // Create notifications
    Create(ctx context.Context, notification Notification) error
    CreateBulk(ctx context.Context, notifications []Notification) error

    // Queries
    ListForUser(ctx context.Context, userID uuid.UUID, filter NotificationFilter) ([]Notification, error)
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)

    // Actions
    MarkRead(ctx context.Context, id uuid.UUID) error
    MarkAllRead(ctx context.Context, userID uuid.UUID) error

    // Dispatch to external channels
    DispatchToChannels(ctx context.Context, teamID uuid.UUID, notification Notification) error
}
```

**Dispatcher:**
- Background goroutine for email/webhook delivery
- Uses channels or worker pool pattern
- Retry logic with exponential backoff
- HMAC-SHA256 signing for webhook payloads

---

### Phase 5: WebSocket Extensions ✅ COMPLETED

**Objective:** Add message types for change management

**Files to modify:**
- `server/entrypoints/ws/messages.go` - Add new message types
- `server/entrypoints/ws/handler.go` - Handle new incoming messages
- `server/entrypoints/ws/hub.go` - Add `BroadcastToAgent(agentID, data)` method

**New Message Types (Server → Agent):**
```go
const (
    TypeChangeApproved   = "change_approved"
    TypeChangeRejected   = "change_rejected"
    TypeExceptionGranted = "exception_granted"
    TypeExceptionDenied  = "exception_denied"
)

type ChangeApprovedPayload struct {
    ChangeID uuid.UUID `json:"change_id"`
    RuleID   uuid.UUID `json:"rule_id"`
}

type ChangeRejectedPayload struct {
    ChangeID     uuid.UUID `json:"change_id"`
    RuleID       uuid.UUID `json:"rule_id"`
    RevertToHash string    `json:"revert_to_hash"`
}

type ExceptionGrantedPayload struct {
    ChangeID    uuid.UUID  `json:"change_id"`
    ExceptionID uuid.UUID  `json:"exception_id"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type ExceptionDeniedPayload struct {
    ChangeID    uuid.UUID `json:"change_id"`
    ExceptionID uuid.UUID `json:"exception_id"`
}
```

**New Message Types (Agent → Server):**
```go
const (
    TypeChangeDetected   = "change_detected"
    TypeChangeUpdated    = "change_updated"
    TypeExceptionRequest = "exception_request"
    TypeRevertComplete   = "revert_complete"
)

type ChangeDetectedPayload struct {
    RuleID          uuid.UUID `json:"rule_id"`
    FilePath        string    `json:"file_path"`
    OriginalHash    string    `json:"original_hash"`
    ModifiedHash    string    `json:"modified_hash"`
    Diff            string    `json:"diff"`
    EnforcementMode string    `json:"enforcement_mode"`
}

type ChangeUpdatedPayload struct {
    ChangeID     uuid.UUID `json:"change_id"`
    ModifiedHash string    `json:"modified_hash"`
    Diff         string    `json:"diff"`
}

type ExceptionRequestPayload struct {
    ChangeID          uuid.UUID `json:"change_id"`
    Justification     string    `json:"justification"`
    ExceptionType     string    `json:"exception_type"`
    RequestedDuration *int      `json:"requested_duration_hours,omitempty"`
}

type RevertCompletePayload struct {
    ChangeID uuid.UUID `json:"change_id"`
}
```

**New Handler Methods:**
```go
type MessageHandler interface {
    // Existing methods...
    HandleHeartbeat(client *Client, payload HeartbeatPayload) error
    HandleDriftReport(client *Client, payload DriftReportPayload) error
    HandleContextDetected(client *Client, payload ContextDetectedPayload) error
    HandleSyncComplete(client *Client, payload SyncCompletePayload) error

    // New methods
    HandleChangeDetected(client *Client, payload ChangeDetectedPayload) error
    HandleChangeUpdated(client *Client, payload ChangeUpdatedPayload) error
    HandleExceptionRequest(client *Client, payload ExceptionRequestPayload) error
    HandleRevertComplete(client *Client, payload RevertCompletePayload) error
}
```

---

### Phase 6: API Handlers ✅ COMPLETED

**Objective:** REST endpoints for web UI

**Files to create:**
- `server/entrypoints/api/handlers/changes.go`
- `server/entrypoints/api/handlers/exceptions.go`
- `server/entrypoints/api/handlers/notifications.go`
- `server/entrypoints/api/handlers/notification_channels.go`

**Extend:**
- `server/entrypoints/api/handlers/rules.go` - Add enforcement mode fields to create/update

**Routes to add in `router.go`:**
```go
// Change Requests
r.Route("/changes", func(r chi.Router) {
    r.Use(middleware.RequirePermission("changes.view"))
    r.Get("/", changesHandler.List)
    r.Get("/{id}", changesHandler.Get)

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequirePermission("changes.approve"))
        r.Post("/{id}/approve", changesHandler.Approve)
        r.Post("/{id}/reject", changesHandler.Reject)
    })
})

// Exceptions
r.Route("/exceptions", func(r chi.Router) {
    r.Use(middleware.RequirePermission("exceptions.view"))
    r.Get("/", exceptionsHandler.List)
    r.Post("/", exceptionsHandler.Create) // User creates appeal

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequirePermission("exceptions.approve"))
        r.Post("/{id}/approve", exceptionsHandler.Approve)
        r.Post("/{id}/deny", exceptionsHandler.Deny)
    })
})

// Notifications
r.Route("/notifications", func(r chi.Router) {
    r.Get("/", notificationsHandler.List)
    r.Get("/unread-count", notificationsHandler.GetUnreadCount)
    r.Post("/{id}/read", notificationsHandler.MarkRead)
    r.Post("/read-all", notificationsHandler.MarkAllRead)
})

// Notification Channels
r.Route("/notification-channels", func(r chi.Router) {
    r.Use(middleware.RequirePermission("notifications.manage"))
    r.Get("/", channelsHandler.List)
    r.Post("/", channelsHandler.Create)
    r.Put("/{id}", channelsHandler.Update)
    r.Delete("/{id}", channelsHandler.Delete)
    r.Post("/{id}/test", channelsHandler.Test)
})
```

---

### Phase 7: Background Worker ✅ COMPLETED

**Objective:** Handle async tasks (email, webhooks, timeouts)

**Files to create:**
- `server/worker/worker.go` - Main worker loop
- `server/worker/timeout_checker.go` - Check expired temporary changes
- `server/worker/email_sender.go` - SMTP email delivery
- `server/worker/webhook_sender.go` - Webhook delivery with HMAC signing

**Worker structure:**
```go
type Worker struct {
    changeService       changes.Service
    notificationService notifications.Service
    wsHub               *ws.Hub

    emailSender   *EmailSender
    webhookSender *WebhookSender

    checkInterval time.Duration
    stopCh        chan struct{}
}

func (w *Worker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-w.stopCh:
            return
        case <-ticker.C:
            w.checkExpiredTemporaryChanges(ctx)
        }
    }
}

func (w *Worker) checkExpiredTemporaryChanges(ctx context.Context) {
    expired, err := w.changeService.HandleExpiredTemporary(ctx)
    if err != nil {
        log.Error().Err(err).Msg("failed to handle expired temporary changes")
        return
    }

    for _, cr := range expired {
        // Send revert command to agent via WebSocket
        w.wsHub.BroadcastToAgent(cr.AgentID, ws.Message{
            Type: ws.TypeChangeRejected,
            Payload: ws.ChangeRejectedPayload{
                ChangeID:     cr.ID,
                RuleID:       cr.RuleID,
                RevertToHash: cr.OriginalHash,
            },
        })

        // Notify user
        w.notificationService.Create(ctx, domain.Notification{
            UserID: cr.UserID,
            TeamID: &cr.TeamID,
            Type:   domain.NotificationTypeChangeAutoReverted,
            Title:  "Change auto-reverted",
            Body:   fmt.Sprintf("Your change to %s was reverted (no approval received)", cr.FilePath),
        })
    }
}
```

**Email Sender:**
```go
type EmailSender struct {
    host     string
    port     int
    username string
    password string
    from     string
}

func (s *EmailSender) Send(ctx context.Context, to []string, subject, body string) error {
    // Standard SMTP implementation with TLS
}
```

**Webhook Sender:**
```go
type WebhookSender struct {
    client     *http.Client
    maxRetries int
}

func (s *WebhookSender) Send(ctx context.Context, url, secret string, payload any) error {
    body, _ := json.Marshal(payload)

    // Sign with HMAC-SHA256
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    signature := hex.EncodeToString(mac.Sum(nil))

    req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Signature-256", "sha256="+signature)

    // Retry with exponential backoff
    return s.sendWithRetry(req)
}
```

**Integration in main.go:**
```go
// Start worker
worker := worker.New(changeService, notificationService, wsHub, workerConfig)
go worker.Start(ctx)
```

---

### Phase 8: Web UI - Notifications Infrastructure ✅ COMPLETED

**Objective:** Client-side notification state and components

**Files to create:**
- `web/src/contexts/NotificationContext.tsx`
- `web/src/components/NotificationBell.tsx`
- `web/src/components/NotificationDropdown.tsx`
- `web/src/domain/notification.ts`

**Domain types:**
```typescript
// web/src/domain/notification.ts
export interface Notification {
  id: string;
  userId: string;
  teamId?: string;
  type: NotificationType;
  title: string;
  body: string;
  metadata: Record<string, any>;
  readAt?: string;
  createdAt: string;
}

export type NotificationType =
  | 'change_detected'
  | 'approval_required'
  | 'change_approved'
  | 'change_rejected'
  | 'change_auto_reverted'
  | 'exception_granted'
  | 'exception_denied';
```

**NotificationContext:**
```typescript
// web/src/contexts/NotificationContext.tsx
interface NotificationContextType {
  notifications: Notification[];
  unreadCount: number;
  loading: boolean;
  error: string | null;
  markRead: (id: string) => Promise<void>;
  markAllRead: () => Promise<void>;
  refetch: () => Promise<void>;
}

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();

  // Fetch notifications on mount and periodically
  useEffect(() => {
    if (!user) return;

    const fetchNotifications = async () => {
      const data = await api.getNotifications();
      setNotifications(data);
      setUnreadCount(data.filter(n => !n.readAt).length);
      setLoading(false);
    };

    fetchNotifications();
    const interval = setInterval(fetchNotifications, 30000); // Poll every 30s
    return () => clearInterval(interval);
  }, [user]);

  // ... markRead, markAllRead, refetch implementations

  return (
    <NotificationContext.Provider value={{ notifications, unreadCount, loading, markRead, markAllRead, refetch }}>
      {children}
    </NotificationContext.Provider>
  );
}
```

**NotificationBell component:**
```typescript
// web/src/components/NotificationBell.tsx
export function NotificationBell() {
  const { unreadCount } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="relative">
      <button onClick={() => setIsOpen(!isOpen)} className="relative p-2">
        <BellIcon className="h-6 w-6" />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>
      {isOpen && <NotificationDropdown onClose={() => setIsOpen(false)} />}
    </div>
  );
}
```

**API additions:**
```typescript
// web/src/lib/api.ts
export async function getNotifications(): Promise<Notification[]> {
  return fetchApi('/notifications');
}

export async function getUnreadCount(): Promise<number> {
  const { count } = await fetchApi('/notifications/unread-count');
  return count;
}

export async function markNotificationRead(id: string): Promise<void> {
  await fetchApi(`/notifications/${id}/read`, { method: 'POST' });
}

export async function markAllNotificationsRead(): Promise<void> {
  await fetchApi('/notifications/read-all', { method: 'POST' });
}
```

---

### Phase 9: Web UI - Changes Page ✅ COMPLETED

**Objective:** Admin interface for change request management

**Files to create:**
- `web/src/app/changes/page.tsx` - Main changes list
- `web/src/app/changes/[id]/page.tsx` - Change detail with diff viewer
- `web/src/app/changes/exceptions/page.tsx` - Exception requests list
- `web/src/components/ChangeRequestTable.tsx`
- `web/src/components/DiffViewer.tsx`
- `web/src/components/ExceptionRequestCard.tsx`
- `web/src/domain/change_request.ts`

**Domain types:**
```typescript
// web/src/domain/change_request.ts
export interface ChangeRequest {
  id: string;
  ruleId: string;
  ruleName: string;
  agentId: string;
  userId: string;
  userName: string;
  teamId: string;
  filePath: string;
  originalHash: string;
  modifiedHash: string;
  diffContent: string;
  status: ChangeRequestStatus;
  enforcementMode: EnforcementMode;
  timeoutAt?: string;
  createdAt: string;
  resolvedAt?: string;
  resolvedByUserId?: string;
  resolvedByUserName?: string;
}

export type ChangeRequestStatus =
  | 'pending'
  | 'approved'
  | 'rejected'
  | 'auto_reverted'
  | 'exception_granted';

export type EnforcementMode = 'block' | 'temporary' | 'warning';

export interface ExceptionRequest {
  id: string;
  changeRequestId: string;
  changeRequest: ChangeRequest;
  userId: string;
  userName: string;
  justification: string;
  exceptionType: 'time_limited' | 'permanent';
  expiresAt?: string;
  status: 'pending' | 'approved' | 'denied';
  createdAt: string;
  resolvedAt?: string;
  resolvedByUserId?: string;
  resolvedByUserName?: string;
}
```

**Changes list page:**
```typescript
// web/src/app/changes/page.tsx
'use client';

export default function ChangesPage() {
  const [changes, setChanges] = useState<ChangeRequest[]>([]);
  const [filter, setFilter] = useState<ChangeRequestFilter>({ status: 'pending' });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getChanges(filter).then(setChanges).finally(() => setLoading(false));
  }, [filter]);

  const handleApprove = async (id: string) => {
    await api.approveChange(id);
    setChanges(changes.filter(c => c.id !== id));
  };

  const handleReject = async (id: string) => {
    await api.rejectChange(id);
    setChanges(changes.filter(c => c.id !== id));
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Change Requests</h1>

      {/* Tabs: Pending, History, Exceptions */}
      <div className="flex gap-4 mb-6">
        <TabLink href="/changes" active={!filter.status || filter.status === 'pending'}>
          Pending
        </TabLink>
        <TabLink href="/changes?status=all">History</TabLink>
        <TabLink href="/changes/exceptions">Exceptions</TabLink>
      </div>

      {/* Filters */}
      <ChangeRequestFilters filter={filter} onChange={setFilter} />

      {/* Table */}
      <ChangeRequestTable
        changes={changes}
        loading={loading}
        onApprove={handleApprove}
        onReject={handleReject}
      />
    </div>
  );
}
```

**DiffViewer component:**
```typescript
// web/src/components/DiffViewer.tsx
import ReactDiffViewer from 'react-diff-viewer-continued';

interface DiffViewerProps {
  oldValue: string;
  newValue: string;
  splitView?: boolean;
}

export function DiffViewer({ oldValue, newValue, splitView = true }: DiffViewerProps) {
  return (
    <ReactDiffViewer
      oldValue={oldValue}
      newValue={newValue}
      splitView={splitView}
      useDarkTheme={true}
      showDiffOnly={false}
    />
  );
}
```

---

### Phase 10: Web UI - Notification Settings ✅ COMPLETED

**Objective:** Configure team notification channels

**Files to create:**
- `web/src/app/settings/channels/page.tsx`
- `web/src/components/ChannelForm.tsx`
- `web/src/components/ChannelList.tsx`
- `web/src/domain/notification_channel.ts`

**Domain types:**
```typescript
// web/src/domain/notification_channel.ts
export interface NotificationChannel {
  id: string;
  teamId: string;
  channelType: 'email' | 'webhook';
  config: EmailConfig | WebhookConfig;
  enabled: boolean;
  createdAt: string;
}

export interface EmailConfig {
  recipients: string[];
  events: string[]; // which notification types to send
}

export interface WebhookConfig {
  url: string;
  secret: string;
  events: string[];
}
```

**Channels page:**
```typescript
// web/src/app/settings/channels/page.tsx
'use client';

export default function NotificationChannelsPage() {
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotificationChannel | null>(null);

  useEffect(() => {
    api.getNotificationChannels().then(setChannels);
  }, []);

  const handleSave = async (channel: Partial<NotificationChannel>) => {
    if (editingChannel) {
      await api.updateNotificationChannel(editingChannel.id, channel);
    } else {
      await api.createNotificationChannel(channel);
    }
    const updated = await api.getNotificationChannels();
    setChannels(updated);
    setShowForm(false);
    setEditingChannel(null);
  };

  const handleTest = async (id: string) => {
    await api.testNotificationChannel(id);
    // Show toast: "Test notification sent"
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Notification Channels</h1>
        <button onClick={() => setShowForm(true)} className="btn-primary">
          Add Channel
        </button>
      </div>

      <ChannelList
        channels={channels}
        onEdit={setEditingChannel}
        onTest={handleTest}
        onToggle={handleToggle}
        onDelete={handleDelete}
      />

      {(showForm || editingChannel) && (
        <ChannelForm
          channel={editingChannel}
          onSave={handleSave}
          onCancel={() => { setShowForm(false); setEditingChannel(null); }}
        />
      )}
    </div>
  );
}
```

---

### Phase 11: Web UI - Rule Editor Enhancement ✅ COMPLETED

**Objective:** Add enforcement mode to rule creation/editing

**Files to modify:**
- `web/src/components/RuleEditor.tsx`

**Changes to RuleEditor:**
```typescript
// Add to existing RuleEditor component

// New state
const [enforcementMode, setEnforcementMode] = useState<EnforcementMode>('block');
const [temporaryTimeoutHours, setTemporaryTimeoutHours] = useState(24);

// Add to form JSX
<div className="mt-6">
  <h3 className="text-lg font-medium mb-4">Enforcement</h3>

  <div className="space-y-4">
    <div>
      <label className="block text-sm font-medium mb-2">
        Enforcement Mode
      </label>
      <select
        value={enforcementMode}
        onChange={(e) => setEnforcementMode(e.target.value as EnforcementMode)}
        className="input w-full"
      >
        <option value="block">Block (default) - Revert immediately, await approval</option>
        <option value="temporary">Temporary - Allow temporarily, auto-revert if not approved</option>
        <option value="warning">Warning - Allow permanently, flag for visibility</option>
      </select>
      <p className="text-sm text-gray-500 mt-1">
        {enforcementMode === 'block' && 'Changes are immediately reverted and require admin approval to apply.'}
        {enforcementMode === 'temporary' && 'Changes apply temporarily but auto-revert if not approved within the timeout.'}
        {enforcementMode === 'warning' && 'Changes apply permanently but are flagged for admin review.'}
      </p>
    </div>

    {enforcementMode === 'temporary' && (
      <div>
        <label className="block text-sm font-medium mb-2">
          Timeout (hours)
        </label>
        <input
          type="number"
          min={1}
          max={168}
          value={temporaryTimeoutHours}
          onChange={(e) => setTemporaryTimeoutHours(parseInt(e.target.value))}
          className="input w-32"
        />
        <p className="text-sm text-gray-500 mt-1">
          Changes will auto-revert after this many hours if not approved.
        </p>
      </div>
    )}
  </div>
</div>
```

---

### Phase 12: Navigation & Polish ✅ COMPLETED

**Objective:** Integrate new pages into navigation

**Files to modify:**
- Navigation component (likely in layout or dedicated nav component)
- `web/src/app/layout.tsx`

**Navigation update:**
```typescript
// Add to navigation items
const navItems = [
  { href: '/', label: 'Dashboard', icon: HomeIcon },
  { href: '/agents', label: 'Agents', icon: ServerIcon },
  { href: '/rules', label: 'Rules', icon: DocumentIcon },
  {
    href: '/changes',
    label: 'Changes',
    icon: DocumentDuplicateIcon,
    badge: pendingChangesCount // From context or API
  },
  { href: '/approvals', label: 'Approvals', icon: CheckCircleIcon },
  // ... admin section
  {
    label: 'Settings',
    children: [
      { href: '/settings/channels', label: 'Notification Channels' },
      // ... other settings
    ]
  }
];

// Add NotificationBell to header
<header className="...">
  <nav>...</nav>
  <div className="flex items-center gap-4">
    <NotificationBell />
    <UserMenu />
  </div>
</header>
```

---

### Phase 13: Testing ⏳ PENDING

**Objective:** Comprehensive test coverage

**Server tests to create:**
```
server/services/changes/service_test.go
server/services/exceptions/service_test.go
server/services/notifications/service_test.go
server/adapters/postgres/change_request_db_test.go
server/adapters/postgres/exception_request_db_test.go
server/adapters/postgres/notification_db_test.go
server/adapters/postgres/notification_channel_db_test.go
server/entrypoints/api/handlers/changes_test.go
server/entrypoints/api/handlers/exceptions_test.go
server/entrypoints/api/handlers/notifications_test.go
server/entrypoints/api/handlers/notification_channels_test.go
server/worker/worker_test.go
```

**Test scenarios:**

**Change Service Tests:**
- Create change request from agent payload
- Approve change updates status and triggers WebSocket message
- Reject change updates status and triggers revert message
- Handle expired temporary changes auto-reverts correctly
- Update from agent updates existing pending request

**Exception Service Tests:**
- Create exception request with valid change_request_id
- Approve exception grants access and notifies user
- Deny exception notifies user
- HasActiveException returns true for valid exception

**Notification Service Tests:**
- Create notification stores correctly
- CreateBulk handles multiple notifications
- MarkRead updates read_at timestamp
- GetUnreadCount returns correct count

**API Handler Tests:**
- List changes with filters
- Approve/reject requires correct permission
- Create exception validates input
- Notification endpoints require authentication

**Worker Tests:**
- Timeout checker finds expired changes
- Email sender retries on failure
- Webhook sender signs payloads correctly

**Web tests to create:**
```
web/src/__tests__/components/NotificationBell.test.tsx
web/src/__tests__/components/ChangeRequestTable.test.tsx
web/src/__tests__/components/DiffViewer.test.tsx
web/src/__tests__/contexts/NotificationContext.test.tsx
web/e2e/changes.spec.ts
web/e2e/notifications.spec.ts
```

---

## Dependencies

| Component | Dependency | Notes |
|-----------|------------|-------|
| Background Worker | goroutine + ticker | Standard library |
| Email Sending | `net/smtp` or `jordan-wright/email` | Standard SMTP, configurable |
| Webhook Signing | `crypto/hmac` + `crypto/sha256` | Standard library for HMAC-SHA256 |
| Diff Generation | `sergi/go-diff` | Unified diff format |
| Desktop Notifications (Agent) | `gen2brain/beeep` | Cross-platform (macOS/Windows/Linux) |
| Diff Viewer (Web) | `react-diff-viewer-continued` | Fork with React 18 support |

---

## Estimated Complexity

| Phase | Complexity | Rationale |
|-------|------------|-----------|
| 1. Database Schema | LOW | Straightforward SQL migrations |
| 2. Domain Models | LOW | Simple struct definitions |
| 3. Repository Layer | MEDIUM | Standard CRUD, but 4 new entities |
| 4. Core Services | HIGH | Complex lifecycle logic, event coordination |
| 5. WebSocket Extensions | MEDIUM | Building on existing infrastructure |
| 6. API Handlers | MEDIUM | Standard REST patterns, many endpoints |
| 7. Background Worker | MEDIUM | New subsystem, retry logic |
| 8. Web Notifications | MEDIUM | New context, real-time updates |
| 9. Changes Page | HIGH | Diff viewer, complex UI interactions |
| 10. Notification Settings | LOW | Standard CRUD form |
| 11. Rule Editor Enhancement | LOW | Adding fields to existing form |
| 12. Navigation & Polish | LOW | UI integration |
| 13. Testing | MEDIUM | Comprehensive coverage needed |

**Overall Complexity: HIGH** - Cross-cutting feature touching all layers

---

## Implementation Order & Dependencies

```
Phase 1 (Schema) ──┬──► Phase 2 (Domain) ──► Phase 3 (Repository) ──┬──► Phase 4 (Services)
                   │                                                  │
                   └──────────────────────────────────────────────────┘
                                          │
                   ┌──────────────────────┼──────────────────────┐
                   ▼                      ▼                      ▼
            Phase 5 (WebSocket)    Phase 6 (API)         Phase 7 (Worker)
                   │                      │                      │
                   └──────────────────────┼──────────────────────┘
                                          │
                   ┌──────────────────────┴──────────────────────┐
                   ▼                                             ▼
            Phase 8 (Web Notifications)                  Phase 11 (Rule Editor)
                   │
         ┌────────┴────────┐
         ▼                 ▼
   Phase 9 (Changes)  Phase 10 (Settings)
         │                 │
         └────────┬────────┘
                  ▼
           Phase 12 (Navigation)
                  │
                  ▼
           Phase 13 (Testing)
```

---

## File Summary

### New Files to Create

**Server (33 files):**
```
server/migrations/
├── 000015_add_change_requests.up.sql
├── 000015_add_change_requests.down.sql
├── 000016_add_exception_requests.up.sql
├── 000016_add_exception_requests.down.sql
├── 000017_add_notifications.up.sql
├── 000017_add_notifications.down.sql
├── 000018_add_notification_channels.up.sql
├── 000018_add_notification_channels.down.sql
├── 000019_add_rule_enforcement.up.sql
├── 000019_add_rule_enforcement.down.sql

server/domain/
├── change_request.go
├── exception_request.go
├── notification.go
├── notification_channel.go

server/adapters/postgres/
├── change_request_db.go
├── exception_request_db.go
├── notification_db.go
├── notification_channel_db.go

server/services/changes/
├── service.go
├── repository.go

server/services/exceptions/
├── service.go
├── repository.go

server/services/notifications/
├── service.go
├── repository.go
├── dispatcher.go

server/worker/
├── worker.go
├── timeout_checker.go
├── email_sender.go
├── webhook_sender.go

server/entrypoints/api/handlers/
├── changes.go
├── exceptions.go
├── notifications.go
├── notification_channels.go
```

**Web (15 files):**
```
web/src/contexts/
├── NotificationContext.tsx

web/src/components/
├── NotificationBell.tsx
├── NotificationDropdown.tsx
├── ChangeRequestTable.tsx
├── DiffViewer.tsx
├── ExceptionRequestCard.tsx
├── ChannelForm.tsx
├── ChannelList.tsx

web/src/app/changes/
├── page.tsx
├── [id]/page.tsx
├── exceptions/page.tsx

web/src/app/settings/channels/
├── page.tsx

web/src/domain/
├── change_request.ts
├── notification.ts
├── notification_channel.ts
```

### Files to Modify

**Server:**
- `server/domain/rule.go` - Add enforcement fields
- `server/entrypoints/ws/messages.go` - Add message types
- `server/entrypoints/ws/handler.go` - Add handlers
- `server/entrypoints/ws/hub.go` - Add BroadcastToAgent
- `server/entrypoints/api/router.go` - Register new routes
- `server/entrypoints/api/handlers/rules.go` - Add enforcement to create/update
- `server/cmd/server/main.go` - Start worker, wire services
- `server/cmd/server/services.go` - Initialize new services

**Web:**
- `web/src/lib/api.ts` - Add API functions
- `web/src/app/providers.tsx` - Add NotificationProvider
- `web/src/app/layout.tsx` - Add NotificationBell to header
- `web/src/components/RuleEditor.tsx` - Add enforcement section
- Navigation component - Add Changes, Notifications links

---

## New Permissions to Add

Add to migration and seed data:

| Code | Description | Category |
|------|-------------|----------|
| `changes.view` | View change requests | changes |
| `changes.approve` | Approve/reject change requests | changes |
| `exceptions.view` | View exception requests | exceptions |
| `exceptions.approve` | Approve/deny exceptions | exceptions |
| `notifications.view` | View own notifications | notifications |
| `notifications.manage` | Manage notification channels | notifications |

---

## Configuration Additions

**Server config (environment variables):**
```bash
# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=notifications@example.com
SMTP_PASSWORD=...
SMTP_FROM=Claudeception <notifications@example.com>

# Worker
WORKER_TIMEOUT_CHECK_INTERVAL=60s
WORKER_RETRY_MAX_ATTEMPTS=5
WORKER_RETRY_INITIAL_DELAY=5s
```

---

## Success Criteria

1. **Change detection works** - Agent detects file modifications and reports to server
2. **Enforcement modes honored** - Block reverts immediately, Temporary allows with timeout, Warning logs only
3. **Admin approval flow** - Web UI shows pending changes, approve/reject works
4. **Exception flow** - Users can appeal, admins can grant time-limited or permanent exceptions
5. **Notifications delivered** - Desktop notifications to users, web/email/webhook to admins
6. **Real-time updates** - WebSocket delivers instant notifications
7. **Resilience** - Offline agents queue changes, replay on reconnect
