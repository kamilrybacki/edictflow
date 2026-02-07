'use client';

import React, { useState, useMemo } from 'react';
import { NotificationChannel, ChannelType } from '@/domain/notification_channel';

interface ChannelFormProps {
  channel?: NotificationChannel | null;
  teamId: string;
  onSave: (data: {
    team_id: string;
    channel_type: string;
    config: Record<string, unknown>;
    enabled: boolean;
  }) => void;
  onCancel: () => void;
}

const notificationEvents = [
  { value: 'change_detected', label: 'Change Detected' },
  { value: 'approval_required', label: 'Approval Required' },
  { value: 'change_approved', label: 'Change Approved' },
  { value: 'change_rejected', label: 'Change Rejected' },
  { value: 'change_auto_reverted', label: 'Change Auto-reverted' },
  { value: 'exception_granted', label: 'Exception Granted' },
  { value: 'exception_denied', label: 'Exception Denied' },
];

export function ChannelForm({ channel, teamId, onSave, onCancel }: ChannelFormProps) {
  const initialValues = useMemo(() => {
    if (!channel) {
      return { recipients: '', webhookUrl: '', webhookSecret: '', events: [] as string[] };
    }
    if (channel.channel_type === 'email') {
      const config = channel.config as { recipients: string[]; events?: string[] };
      return {
        recipients: config.recipients?.join(', ') || '',
        webhookUrl: '',
        webhookSecret: '',
        events: config.events || [],
      };
    }
    const config = channel.config as { url: string; secret?: string; events?: string[] };
    return {
      recipients: '',
      webhookUrl: config.url || '',
      webhookSecret: config.secret || '',
      events: config.events || [],
    };
  }, [channel]);

  const [channelType, setChannelType] = useState<ChannelType>(
    channel?.channel_type || 'email'
  );
  const [enabled, setEnabled] = useState(channel?.enabled ?? true);
  const [recipients, setRecipients] = useState(initialValues.recipients);
  const [webhookUrl, setWebhookUrl] = useState(initialValues.webhookUrl);
  const [webhookSecret, setWebhookSecret] = useState(initialValues.webhookSecret);
  const [selectedEvents, setSelectedEvents] = useState<string[]>(initialValues.events);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    let config: Record<string, unknown>;
    if (channelType === 'email') {
      config = {
        recipients: recipients.split(',').map((r) => r.trim()).filter(Boolean),
        events: selectedEvents.length > 0 ? selectedEvents : undefined,
      };
    } else {
      config = {
        url: webhookUrl,
        secret: webhookSecret || undefined,
        events: selectedEvents.length > 0 ? selectedEvents : undefined,
      };
    }

    onSave({
      team_id: teamId,
      channel_type: channelType,
      config,
      enabled,
    });
  };

  const toggleEvent = (event: string) => {
    setSelectedEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event]
    );
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-lg w-full max-w-lg p-6">
        <h2 className="text-xl font-bold text-white mb-6">
          {channel ? 'Edit Channel' : 'Add Notification Channel'}
        </h2>

        <form onSubmit={handleSubmit}>
          {/* Channel Type */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Channel Type
            </label>
            <select
              value={channelType}
              onChange={(e) => setChannelType(e.target.value as ChannelType)}
              disabled={!!channel}
              className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white disabled:opacity-50"
            >
              <option value="email">Email</option>
              <option value="webhook">Webhook</option>
            </select>
          </div>

          {/* Email Config */}
          {channelType === 'email' && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Recipients (comma-separated)
              </label>
              <input
                type="text"
                value={recipients}
                onChange={(e) => setRecipients(e.target.value)}
                placeholder="admin@example.com, ops@example.com"
                className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-500"
              />
            </div>
          )}

          {/* Webhook Config */}
          {channelType === 'webhook' && (
            <>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Webhook URL
                </label>
                <input
                  type="url"
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder="https://hooks.example.com/notify"
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-500"
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Secret (for HMAC signing)
                </label>
                <input
                  type="password"
                  value={webhookSecret}
                  onChange={(e) => setWebhookSecret(e.target.value)}
                  placeholder="Optional secret for signature verification"
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white placeholder-gray-500"
                />
              </div>
            </>
          )}

          {/* Event Filters */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Notify on events (leave empty for all)
            </label>
            <div className="flex flex-wrap gap-2">
              {notificationEvents.map((event) => (
                <button
                  key={event.value}
                  type="button"
                  onClick={() => toggleEvent(event.value)}
                  className={`px-3 py-1 text-sm rounded border ${
                    selectedEvents.includes(event.value)
                      ? 'bg-blue-600 border-blue-500 text-white'
                      : 'bg-gray-700 border-gray-600 text-gray-300 hover:border-gray-500'
                  }`}
                >
                  {event.label}
                </button>
              ))}
            </div>
          </div>

          {/* Enabled Toggle */}
          <div className="mb-6">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="w-4 h-4 rounded border-gray-600 text-blue-600 focus:ring-blue-500"
              />
              <span className="text-sm text-gray-300">Enabled</span>
            </label>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 text-gray-300 hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium"
            >
              {channel ? 'Save Changes' : 'Create Channel'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
