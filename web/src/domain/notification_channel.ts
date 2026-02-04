export type ChannelType = 'email' | 'webhook';

export interface EmailConfig {
  recipients: string[];
  events?: string[];
}

export interface WebhookConfig {
  url: string;
  secret?: string;
  events?: string[];
}

export interface NotificationChannel {
  id: string;
  team_id: string;
  channel_type: ChannelType;
  config: EmailConfig | WebhookConfig;
  enabled: boolean;
  created_at: string;
}
