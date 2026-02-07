import { Notification } from '@/domain/notification';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchNotifications(unread?: boolean): Promise<Notification[]> {
  let url = `${getApiUrlCached()}/api/v1/notifications/`;
  if (unread) {
    url += '?unread=true';
  }
  const res = await fetch(url, { headers: getAuthHeaders() });
  if (!res.ok) {
    throw new Error(`Failed to fetch notifications: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function fetchUnreadCount(): Promise<number> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notifications/unread-count`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch unread count: ${res.statusText}`);
  }
  const data = await res.json();
  return data?.count ?? 0;
}

export async function markNotificationRead(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notifications/${id}/read`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to mark notification as read: ${res.statusText}`);
  }
}

export async function markAllNotificationsRead(): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notifications/read-all`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to mark all notifications as read: ${res.statusText}`);
  }
}
