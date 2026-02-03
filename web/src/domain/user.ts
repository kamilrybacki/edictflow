export type AuthProvider = 'github' | 'gitlab' | 'google' | 'local';
export type Role = 'admin' | 'member';

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string;
  authProvider: AuthProvider;
  role: Role;
  teamId: string;
  createdAt: string;
}
