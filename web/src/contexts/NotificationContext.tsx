'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { NotificationState } from '@/domain/notification';
import {
  fetchNotifications,
  fetchUnreadCount,
  markNotificationRead,
  markAllNotificationsRead,
} from '@/lib/api';
import { useAuth } from './AuthContext';

interface NotificationContextType extends NotificationState {
  markRead: (id: string) => Promise<void>;
  markAllRead: () => Promise<void>;
  refetch: () => Promise<void>;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

const POLL_INTERVAL = 30000; // 30 seconds

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<NotificationState>({
    notifications: [],
    unreadCount: 0,
    loading: true,
    error: null,
  });

  const { isAuthenticated, isLoading: authLoading } = useAuth();

  const fetchData = useCallback(async () => {
    if (!isAuthenticated) {
      setState({ notifications: [], unreadCount: 0, loading: false, error: null });
      return;
    }

    try {
      const [notifications, unreadCount] = await Promise.all([
        fetchNotifications(),
        fetchUnreadCount(),
      ]);
      setState({
        notifications,
        unreadCount,
        loading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        loading: false,
        error: err instanceof Error ? err.message : 'Failed to fetch notifications',
      }));
    }
  }, [isAuthenticated]);

  // Initial fetch and polling
  useEffect(() => {
    if (authLoading) return;

    fetchData();

    if (isAuthenticated) {
      const interval = setInterval(fetchData, POLL_INTERVAL);
      return () => clearInterval(interval);
    }
  }, [isAuthenticated, authLoading, fetchData]);

  const markRead = useCallback(async (id: string) => {
    await markNotificationRead(id);
    setState((prev) => ({
      ...prev,
      notifications: prev.notifications.map((n) =>
        n.id === id ? { ...n, read_at: new Date().toISOString() } : n
      ),
      unreadCount: Math.max(0, prev.unreadCount - 1),
    }));
  }, []);

  const markAllRead = useCallback(async () => {
    await markAllNotificationsRead();
    setState((prev) => ({
      ...prev,
      notifications: prev.notifications.map((n) => ({
        ...n,
        read_at: n.read_at || new Date().toISOString(),
      })),
      unreadCount: 0,
    }));
  }, []);

  const refetch = useCallback(async () => {
    await fetchData();
  }, [fetchData]);

  return (
    <NotificationContext.Provider
      value={{
        ...state,
        markRead,
        markAllRead,
        refetch,
      }}
    >
      {children}
    </NotificationContext.Provider>
  );
}

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
}
