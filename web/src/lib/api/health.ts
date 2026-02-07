import { getApiUrlCached } from './http';

export async function fetchServiceInfo(): Promise<{ service: string; version: string; status: string } | null> {
  try {
    const res = await fetch(`${getApiUrlCached()}/`);
    if (res.ok) {
      return res.json();
    }
    return null;
  } catch {
    return null;
  }
}
