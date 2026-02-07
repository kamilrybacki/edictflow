// Core HTTP utilities for API calls

// API URL is determined at runtime in the browser
function getApiUrl(): string {
  if (typeof window !== 'undefined') {
    // Client-side: check for runtime config
    const runtimeUrl = (window as unknown as { __API_URL__?: string }).__API_URL__;
    if (runtimeUrl) return runtimeUrl;

    // Use environment variable if available (set at build time)
    if (process.env.NEXT_PUBLIC_API_URL) {
      return process.env.NEXT_PUBLIC_API_URL;
    }

    // Default fallback - assume API is on port 8080
    return `http://${window.location.hostname}:8080`;
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
}

// Cache the API URL
let cachedApiUrl: string | null = null;
export function getApiUrlCached(): string {
  if (!cachedApiUrl) {
    cachedApiUrl = getApiUrl();
  }
  return cachedApiUrl;
}

export const TOKEN_KEY = 'auth_token';

// Get auth headers with token from localStorage
export function getAuthHeaders(): HeadersInit {
  const token = typeof window !== 'undefined' ? localStorage.getItem(TOKEN_KEY) : null;
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}
