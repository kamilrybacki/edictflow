import { NotificationChannel } from '@/domain/notification_channel';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchNotificationChannels(teamId: string): Promise<NotificationChannel[]> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notification-channels/?team_id=${teamId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch notification channels: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export async function createNotificationChannel(data: {
  team_id: string;
  channel_type: string;
  config: Record<string, unknown>;
  enabled: boolean;
}): Promise<NotificationChannel> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notification-channels/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    throw new Error(`Failed to create notification channel: ${res.statusText}`);
  }
  return res.json();
}

export async function updateNotificationChannel(
  id: string,
  data: {
    channel_type: string;
    config: Record<string, unknown>;
    enabled: boolean;
  }
): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notification-channels/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    throw new Error(`Failed to update notification channel: ${res.statusText}`);
  }
}

export async function deleteNotificationChannel(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notification-channels/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to delete notification channel: ${res.statusText}`);
  }
}

export async function testNotificationChannel(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/notification-channels/${id}/test`, {
    method: 'POST',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to test notification channel: ${res.statusText}`);
  }
}
