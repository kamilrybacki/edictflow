export type ChangeRequestStatus =
  | 'pending'
  | 'approved'
  | 'rejected'
  | 'auto_reverted'
  | 'exception_granted';

export type EnforcementMode = 'block' | 'temporary' | 'warning';

export interface ChangeRequest {
  id: string;
  rule_id: string;
  agent_id: string;
  user_id: string;
  team_id: string;
  file_path: string;
  original_hash: string;
  modified_hash: string;
  diff_content: string;
  status: ChangeRequestStatus;
  enforcement_mode: EnforcementMode;
  timeout_at?: string;
  created_at: string;
  resolved_at?: string;
  resolved_by_user_id?: string;
}

export type ExceptionRequestStatus = 'pending' | 'approved' | 'denied';

export type ExceptionType = 'time_limited' | 'permanent';

export interface ExceptionRequest {
  id: string;
  change_request_id: string;
  user_id: string;
  justification: string;
  exception_type: ExceptionType;
  expires_at?: string;
  status: ExceptionRequestStatus;
  created_at: string;
  resolved_at?: string;
  resolved_by_user_id?: string;
}
