'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { NotificationChannel } from '@/domain/notification_channel';
import { ChannelList } from '@/components/ChannelList';
import { ChannelForm } from '@/components/ChannelForm';
import { useAuth, useRequirePermission } from '@/contexts/AuthContext';
import { NotificationBell } from '@/components/NotificationBell';
import { UserMenu } from '@/components/UserMenu';
import {
  fetchNotificationChannels,
  createNotificationChannel,
  updateNotificationChannel,
  deleteNotificationChannel,
  testNotificationChannel,
} from '@/lib/api';

export default function NotificationChannelsPage() {
  const auth = useRequirePermission('notifications.manage');
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotificationChannel | null>(null);
  const [testMessage, setTestMessage] = useState<string | null>(null);

  const teamId = auth.user?.teamId;

  useEffect(() => {
    if (!teamId) return;

    async function loadChannels(tid: string) {
      setLoading(true);
      try {
        const data = await fetchNotificationChannels(tid);
        setChannels(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load channels');
      } finally {
        setLoading(false);
      }
    }

    loadChannels(teamId);
  }, [teamId]);

  const handleSave = async (data: {
    team_id: string;
    channel_type: string;
    config: Record<string, unknown>;
    enabled: boolean;
  }) => {
    try {
      if (editingChannel) {
        await updateNotificationChannel(editingChannel.id, {
          channel_type: data.channel_type,
          config: data.config,
          enabled: data.enabled,
        });
      } else {
        await createNotificationChannel(data);
      }
      const updated = await fetchNotificationChannels(teamId!);
      setChannels(updated);
      setShowForm(false);
      setEditingChannel(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save channel');
    }
  };

  const handleTest = async (id: string) => {
    try {
      await testNotificationChannel(id);
      setTestMessage('Test notification sent successfully');
      setTimeout(() => setTestMessage(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send test');
    }
  };

  const handleToggle = async (id: string, enabled: boolean) => {
    const channel = channels.find((c) => c.id === id);
    if (!channel) return;

    try {
      await updateNotificationChannel(id, {
        channel_type: channel.channel_type,
        config: channel.config as unknown as Record<string, unknown>,
        enabled,
      });
      setChannels(
        channels.map((c) => (c.id === id ? { ...c, enabled } : c))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to toggle channel');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this channel?')) return;

    try {
      await deleteNotificationChannel(id);
      setChannels(channels.filter((c) => c.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete channel');
    }
  };

  if (auth.isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link href="/" className="text-xl font-bold">
                Edictflow
              </Link>
              <span className="text-gray-500">/</span>
              <span className="text-gray-300">Settings</span>
              <span className="text-gray-500">/</span>
              <span className="text-gray-300">Notification Channels</span>
            </div>
            <div className="flex items-center gap-4">
              <NotificationBell />
              <UserMenu />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-white">Notification Channels</h1>
          <button
            onClick={() => {
              setEditingChannel(null);
              setShowForm(true);
            }}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium"
          >
            Add Channel
          </button>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-900/20 border border-red-500/50 rounded-lg text-red-400">
            {error}
            <button
              onClick={() => setError(null)}
              className="ml-4 text-sm underline"
            >
              Dismiss
            </button>
          </div>
        )}

        {testMessage && (
          <div className="mb-4 p-4 bg-green-900/20 border border-green-500/50 rounded-lg text-green-400">
            {testMessage}
          </div>
        )}

        <div className="bg-gray-800 rounded-lg overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
            </div>
          ) : (
            <ChannelList
              channels={channels}
              onEdit={(channel) => {
                setEditingChannel(channel);
                setShowForm(true);
              }}
              onTest={handleTest}
              onToggle={handleToggle}
              onDelete={handleDelete}
            />
          )}
        </div>
      </main>

      {/* Form Modal */}
      {showForm && teamId && (
        <ChannelForm
          channel={editingChannel}
          teamId={teamId}
          onSave={handleSave}
          onCancel={() => {
            setShowForm(false);
            setEditingChannel(null);
          }}
        />
      )}
    </div>
  );
}
