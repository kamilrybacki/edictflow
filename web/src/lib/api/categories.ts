import { Category } from '@/domain/rule';
import { getApiUrlCached, getAuthHeaders } from './http';

export async function fetchCategories(orgId?: string): Promise<Category[]> {
  const params = orgId ? `?org_id=${orgId}` : '';
  const res = await fetch(`${getApiUrlCached()}/api/v1/categories/${params}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch categories: ${res.statusText}`);
  }
  const data = await res.json();
  return data || [];
}

export interface CreateCategoryRequest {
  name: string;
  org_id?: string;
  display_order: number;
}

export async function createCategory(data: CreateCategoryRequest): Promise<Category> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/categories/`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    throw new Error(`Failed to create category: ${res.statusText}`);
  }
  return res.json();
}

export async function deleteCategory(id: string): Promise<void> {
  const res = await fetch(`${getApiUrlCached()}/api/v1/categories/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    throw new Error(`Failed to delete category: ${res.statusText}`);
  }
}
