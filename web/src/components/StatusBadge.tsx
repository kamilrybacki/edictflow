'use client';

import { useEffect, useState } from 'react';
import { fetchServiceInfo } from '@/lib/api';

export function StatusBadge() {
  const [status, setStatus] = useState<'loading' | 'connected' | 'disconnected'>('loading');
  const [info, setInfo] = useState<{ service: string; version: string } | null>(null);

  useEffect(() => {
    const checkStatus = async () => {
      const serviceInfo = await fetchServiceInfo();
      if (serviceInfo) {
        setStatus('connected');
        setInfo({ service: serviceInfo.service, version: serviceInfo.version });
      } else {
        setStatus('disconnected');
        setInfo(null);
      }
    };

    checkStatus();
    const interval = setInterval(checkStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex items-center gap-2">
      <div
        className={`h-2 w-2 rounded-full ${
          status === 'connected'
            ? 'bg-green-500'
            : status === 'disconnected'
            ? 'bg-red-500'
            : 'bg-yellow-500 animate-pulse'
        }`}
      />
      <span className="text-sm text-zinc-600 dark:text-zinc-400">
        {status === 'connected' && info
          ? `${info.service} v${info.version}`
          : status === 'disconnected'
          ? 'API Disconnected'
          : 'Connecting...'}
      </span>
    </div>
  );
}
