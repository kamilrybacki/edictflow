export type NotificationType =
  | 'change_detected'
  | 'approval_required'
  | 'change_approved'
  | 'change_rejected'
  | 'change_auto_reverted'
  | 'exception_granted'
  | 'exception_denied';

export interface Notification {
  id: string;
  user_id: string;
  team_id?: string;
  type: NotificationType;
  title: string;
  body: string;
  metadata: Record<string, unknown>;
  read_at?: string;
  created_at: string;
}

export interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  loading: boolean;
  error: string | null;
}
