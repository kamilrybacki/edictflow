import { LoginRequest, RegisterRequest, AuthResponse } from '@/domain/user';
import { getApiUrlCached } from './http';

export async function login(request: LoginRequest): Promise<AuthResponse> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || 'Login failed');
  }
  return res.json();
}

export async function register(request: RegisterRequest): Promise<AuthResponse> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || 'Registration failed');
  }
  return res.json();
}
