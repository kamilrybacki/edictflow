export type AuthProvider = 'github' | 'gitlab' | 'google' | 'local';

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string;
  authProvider: AuthProvider;
  teamId?: string;
  permissions: string[];
  isActive: boolean;
  createdAt: string;
  lastLoginAt?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  name: string;
  password: string;
  teamId?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}
