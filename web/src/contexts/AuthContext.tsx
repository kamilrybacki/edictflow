'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { User, AuthState, LoginRequest, RegisterRequest } from '@/domain/user';
import { login as apiLogin, register as apiRegister } from '@/lib/api';

interface AuthContextType extends AuthState {
  login: (request: LoginRequest) => Promise<void>;
  register: (request: RegisterRequest) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
  hasAnyPermission: (...permissions: string[]) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

function parseJwt(token: string): { sub: string; email: string; team_id?: string; permissions: string[]; exp: number } | null {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  } catch {
    return null;
  }
}

function isTokenExpired(token: string): boolean {
  const payload = parseJwt(token);
  if (!payload) return true;
  return Date.now() >= payload.exp * 1000;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: true,
  });

  // Load auth state from localStorage on mount
  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY);
    const userJson = localStorage.getItem(USER_KEY);

    if (token && userJson) {
      if (isTokenExpired(token)) {
        // Token expired, clear storage
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
        setState({ user: null, token: null, isAuthenticated: false, isLoading: false });
      } else {
        const user = JSON.parse(userJson) as User;
        setState({ user, token, isAuthenticated: true, isLoading: false });
      }
    } else {
      setState({ user: null, token: null, isAuthenticated: false, isLoading: false });
    }
  }, []);

  const login = useCallback(async (request: LoginRequest) => {
    const response = await apiLogin(request);
    localStorage.setItem(TOKEN_KEY, response.token);
    localStorage.setItem(USER_KEY, JSON.stringify(response.user));
    setState({
      user: response.user,
      token: response.token,
      isAuthenticated: true,
      isLoading: false,
    });
  }, []);

  const register = useCallback(async (request: RegisterRequest) => {
    const response = await apiRegister(request);
    localStorage.setItem(TOKEN_KEY, response.token);
    localStorage.setItem(USER_KEY, JSON.stringify(response.user));
    setState({
      user: response.user,
      token: response.token,
      isAuthenticated: true,
      isLoading: false,
    });
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setState({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
    });
  }, []);

  const hasPermission = useCallback(
    (permission: string) => {
      return state.user?.permissions.includes(permission) ?? false;
    },
    [state.user]
  );

  const hasAnyPermission = useCallback(
    (...permissions: string[]) => {
      if (!state.user) return false;
      return permissions.some((p) => state.user!.permissions.includes(p));
    },
    [state.user]
  );

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        register,
        logout,
        hasPermission,
        hasAnyPermission,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function useRequireAuth() {
  const auth = useAuth();

  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated) {
      window.location.href = '/login';
    }
  }, [auth.isLoading, auth.isAuthenticated]);

  return auth;
}

export function useRequirePermission(permission: string) {
  const auth = useRequireAuth();

  useEffect(() => {
    if (!auth.isLoading && auth.isAuthenticated && !auth.hasPermission(permission)) {
      window.location.href = '/';
    }
  }, [auth.isLoading, auth.isAuthenticated, auth.hasPermission, permission]);

  return auth;
}
