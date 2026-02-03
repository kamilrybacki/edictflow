package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockPermissionProvider struct {
	permissions map[string][]string
	called      bool
}

func (m *mockPermissionProvider) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	m.called = true
	return m.permissions[userID], nil
}

func TestRequirePermission_UsesContextPermissions(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{},
	}

	pm := NewPermission(provider)
	middleware := pm.RequirePermission("read_rules")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with user and permissions in context (simulating JWT auth)
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	ctx = context.WithValue(ctx, permissionsKey, []string{"read_rules", "create_rules"})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Provider should not have been called since permissions were in context
	if provider.called {
		t.Error("expected provider to NOT be called when permissions are in context")
	}
}

func TestRequirePermission_Allowed(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{
			"user-1": {"read_rules", "create_rules"},
		},
	}

	pm := NewPermission(provider)
	middleware := pm.RequirePermission("read_rules")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with user in context
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRequirePermission_Denied(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{
			"user-1": {"read_rules"},
		},
	}

	pm := NewPermission(provider)
	middleware := pm.RequirePermission("delete_rules")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("DELETE", "/rules/1", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequirePermission_NoUser(t *testing.T) {
	provider := &mockPermissionProvider{}

	pm := NewPermission(provider)
	middleware := pm.RequirePermission("read_rules")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAnyPermission_Allowed(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{
			"user-1": {"approve_local"},
		},
	}

	pm := NewPermission(provider)
	middleware := pm.RequireAnyPermission("approve_local", "approve_project", "approve_global")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/approve", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireAnyPermission_Denied(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{
			"user-1": {"read_rules"},
		},
	}

	pm := NewPermission(provider)
	middleware := pm.RequireAnyPermission("approve_local", "approve_project")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/approve", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAllPermissions_Allowed(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{
			"user-1": {"read_rules", "create_rules", "delete_rules"},
		},
	}

	pm := NewPermission(provider)
	middleware := pm.RequireAllPermissions("read_rules", "create_rules")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/rules", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireAllPermissions_Denied(t *testing.T) {
	provider := &mockPermissionProvider{
		permissions: map[string][]string{
			"user-1": {"read_rules"},
		},
	}

	pm := NewPermission(provider)
	middleware := pm.RequireAllPermissions("read_rules", "create_rules")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/rules", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user-1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
