'use client';

import React, { memo, useCallback } from 'react';
import { NotificationChannel, ChannelType } from '@/domain/notification_channel';

interface ChannelListProps {
  channels: NotificationChannel[];
  onEdit: (channel: NotificationChannel) => void;
  onTest: (id: string) => void;
  onToggle: (id: string, enabled: boolean) => void;
  onDelete: (id: string) => void;
}

const channelTypeIcons: Record<ChannelType, string> = {
  email: '📧',
  webhook: '🔗',
};

const channelTypeLabels: Record<ChannelType, string> = {
  email: 'Email',
  webhook: 'Webhook',
};

// Memoized channel item to prevent re-renders
const ChannelItem = memo(function ChannelItem({
  channel,
  onEdit,
  onTest,
  onToggle,
  onDelete,
}: {
  channel: NotificationChannel;
  onEdit: (channel: NotificationChannel) => void;
  onTest: (id: string) => void;
  onToggle: (id: string, enabled: boolean) => void;
  onDelete: (id: string) => void;
}) {
  const handleTest = useCallback(() => onTest(channel.id), [onTest, channel.id]);
  const handleToggle = useCallback(() => onToggle(channel.id, !channel.enabled), [onToggle, channel.id, channel.enabled]);
  const handleEdit = useCallback(() => onEdit(channel), [onEdit, channel]);
  const handleDelete = useCallback(() => onDelete(channel.id), [onDelete, channel.id]);

  return (
    <div className="p-4 flex items-center justify-between hover:bg-gray-700/50">
      <div className="flex items-center gap-4">
        <span className="text-2xl">{channelTypeIcons[channel.channel_type]}</span>
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium text-white">
              {channelTypeLabels[channel.channel_type]}
            </span>
            {!channel.enabled && (
              <span className="px-2 py-0.5 text-xs bg-gray-600 text-gray-300 rounded">
                Disabled
              </span>
            )}
          </div>
          <div className="text-sm text-gray-400">
            {channel.channel_type === 'email' && (
              <>
                {(channel.config as { recipients: string[] }).recipients?.length || 0}{' '}
                recipients
              </>
            )}
            {channel.channel_type === 'webhook' && (
              <span className="font-mono">
                {(channel.config as { url: string }).url}
              </span>
            )}
          </div>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <button
          onClick={handleTest}
          className="px-3 py-1.5 text-sm text-gray-300 hover:text-white bg-gray-700 hover:bg-gray-600 rounded"
        >
          Test
        </button>
        <button
          onClick={handleToggle}
          className={`px-3 py-1.5 text-sm rounded ${
            channel.enabled
              ? 'text-yellow-400 hover:text-yellow-300 bg-yellow-900/30 hover:bg-yellow-900/50'
              : 'text-green-400 hover:text-green-300 bg-green-900/30 hover:bg-green-900/50'
          }`}
        >
          {channel.enabled ? 'Disable' : 'Enable'}
        </button>
        <button
          onClick={handleEdit}
          className="px-3 py-1.5 text-sm text-blue-400 hover:text-blue-300 bg-blue-900/30 hover:bg-blue-900/50 rounded"
        >
          Edit
        </button>
        <button
          onClick={handleDelete}
          className="px-3 py-1.5 text-sm text-red-400 hover:text-red-300 bg-red-900/30 hover:bg-red-900/50 rounded"
        >
          Delete
        </button>
      </div>
    </div>
  );
});

export const ChannelList = memo(function ChannelList({
  channels,
  onEdit,
  onTest,
  onToggle,
  onDelete,
}: ChannelListProps) {
  if (channels.length === 0) {
    return (
      <div className="text-center py-12 text-gray-400">
        No notification channels configured
      </div>
    );
  }

  return (
    <div className="divide-y divide-gray-700">
      {channels.map((channel) => (
        <ChannelItem
          key={channel.id}
          channel={channel}
          onEdit={onEdit}
          onTest={onTest}
          onToggle={onToggle}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
});
