package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
)

type mockRolesService struct {
	roles       map[string]domain.RoleEntity
	permissions map[string][]domain.Permission
	userRoles   map[string][]string // userID -> roleIDs
}

func newMockRolesService() *mockRolesService {
	return &mockRolesService{
		roles:       make(map[string]domain.RoleEntity),
		permissions: make(map[string][]domain.Permission),
		userRoles:   make(map[string][]string),
	}
}

func (m *mockRolesService) Create(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.RoleEntity, error) {
	role := domain.NewRoleEntity(name, description, hierarchyLevel, parentRoleID, teamID)
	m.roles[role.ID] = role
	return role, nil
}

func (m *mockRolesService) GetByID(ctx context.Context, id string) (domain.RoleEntity, error) {
	if role, ok := m.roles[id]; ok {
		return role, nil
	}
	return domain.RoleEntity{}, errors.New("role not found")
}

func (m *mockRolesService) List(ctx context.Context, teamID *string) ([]domain.RoleEntity, error) {
	var result []domain.RoleEntity
	for _, r := range m.roles {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockRolesService) Update(ctx context.Context, role domain.RoleEntity) error {
	if _, ok := m.roles[role.ID]; !ok {
		return errors.New("role not found")
	}
	m.roles[role.ID] = role
	return nil
}

func (m *mockRolesService) Delete(ctx context.Context, id string) error {
	if _, ok := m.roles[id]; !ok {
		return errors.New("role not found")
	}
	delete(m.roles, id)
	return nil
}

func (m *mockRolesService) GetRoleWithPermissions(ctx context.Context, id string) (domain.RoleEntity, error) {
	role, ok := m.roles[id]
	if !ok {
		return domain.RoleEntity{}, errors.New("role not found")
	}
	role.Permissions = m.permissions[id]
	return role, nil
}

func (m *mockRolesService) AddPermission(ctx context.Context, roleID, permissionID string) error {
	m.permissions[roleID] = append(m.permissions[roleID], domain.Permission{ID: permissionID, Code: permissionID})
	return nil
}

func (m *mockRolesService) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	perms := m.permissions[roleID]
	for i, p := range perms {
		if p.ID == permissionID {
			m.permissions[roleID] = append(perms[:i], perms[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockRolesService) AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error {
	m.userRoles[userID] = append(m.userRoles[userID], roleID)
	return nil
}

func (m *mockRolesService) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	roles := m.userRoles[userID]
	for i, r := range roles {
		if r == roleID {
			m.userRoles[userID] = append(roles[:i], roles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockRolesService) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return []domain.Permission{
		{ID: "perm-1", Code: "rules:read", Description: "Read rules"},
		{ID: "perm-2", Code: "rules:write", Description: "Write rules"},
	}, nil
}

func TestRolesHandler_Create(t *testing.T) {
	svc := newMockRolesService()
	h := handlers.NewRolesHandler(svc)

	body := `{"name":"Admin","description":"Administrator role","hierarchy_level":1}`
	req := httptest.NewRequest("POST", "/roles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handlers.RoleResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "Admin" {
		t.Errorf("expected name 'Admin', got '%s'", resp.Name)
	}
}

func TestRolesHandler_List(t *testing.T) {
	svc := newMockRolesService()
	svc.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin"}
	svc.roles["role-2"] = domain.RoleEntity{ID: "role-2", Name: "Member"}

	h := handlers.NewRolesHandler(svc)
	req := httptest.NewRequest("GET", "/roles", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp []handlers.RoleResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 roles, got %d", len(resp))
	}
}

func TestRolesHandler_Get(t *testing.T) {
	svc := newMockRolesService()
	svc.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", Description: "Administrator"}

	h := handlers.NewRolesHandler(svc)

	r := chi.NewRouter()
	r.Get("/roles/{id}", h.Get)

	req := httptest.NewRequest("GET", "/roles/role-1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handlers.RoleDetailResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.ID != "role-1" {
		t.Errorf("expected ID 'role-1', got '%s'", resp.ID)
	}
}

func TestRolesHandler_AddPermission(t *testing.T) {
	svc := newMockRolesService()
	svc.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin"}

	h := handlers.NewRolesHandler(svc)

	r := chi.NewRouter()
	r.Post("/roles/{id}/permissions", h.AddPermission)

	body := `{"permission_id":"perm-1"}`
	req := httptest.NewRequest("POST", "/roles/role-1/permissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(svc.permissions["role-1"]) != 1 {
		t.Error("expected permission to be added")
	}
}

func TestRolesHandler_Delete(t *testing.T) {
	svc := newMockRolesService()
	svc.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin"}

	h := handlers.NewRolesHandler(svc)

	r := chi.NewRouter()
	r.Delete("/roles/{id}", h.Delete)

	req := httptest.NewRequest("DELETE", "/roles/role-1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if _, ok := svc.roles["role-1"]; ok {
		t.Error("expected role to be deleted")
	}
}

func TestRolesHandler_AssignUser(t *testing.T) {
	svc := newMockRolesService()
	svc.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin"}

	h := handlers.NewRolesHandler(svc)

	r := chi.NewRouter()
	r.Post("/roles/{id}/users", h.AssignUser)

	body := `{"user_id":"user-1"}`
	req := httptest.NewRequest("POST", "/roles/role-1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(svc.userRoles["user-1"]) != 1 {
		t.Error("expected user to be assigned to role")
	}
}

func TestRolesHandler_RemoveUser(t *testing.T) {
	svc := newMockRolesService()
	svc.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin"}
	svc.userRoles["user-1"] = []string{"role-1"}

	h := handlers.NewRolesHandler(svc)

	r := chi.NewRouter()
	r.Delete("/roles/{id}/users/{userId}", h.RemoveUser)

	req := httptest.NewRequest("DELETE", "/roles/role-1/users/user-1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(svc.userRoles["user-1"]) != 0 {
		t.Error("expected user to be removed from role")
	}
}

func TestRolesHandler_ListPermissions(t *testing.T) {
	svc := newMockRolesService()
	h := handlers.NewRolesHandler(svc)

	req := httptest.NewRequest("GET", "/roles/permissions", nil)
	rec := httptest.NewRecorder()

	h.ListPermissions(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []handlers.PermissionResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(resp))
	}
}
