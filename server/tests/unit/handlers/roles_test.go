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
	"github.com/kamilrybacki/edictflow/server/tests/testutil"
)

func TestRolesHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful creation",
			body:           `{"name":"Admin","description":"Administrator role","hierarchy_level":100}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.RoleResponse
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Name != "Admin" {
					t.Errorf("expected name 'Admin', got '%s'", resp.Name)
				}
			},
		},
		{
			name:           "invalid JSON",
			body:           `{invalid}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			body: `{"description":"Test role","hierarchy_level":50}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					if name == "" {
						return domain.Role{}, errors.New("name is required")
					}
					return domain.NewRole(name, description, hierarchyLevel, parentRoleID, teamID), nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing hierarchy level",
			body: `{"name":"Test","description":"Test role"}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					if hierarchyLevel == 0 {
						return domain.Role{}, errors.New("hierarchy level is required")
					}
					return domain.NewRole(name, description, hierarchyLevel, parentRoleID, teamID), nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative hierarchy level",
			body: `{"name":"Test","description":"Test role","hierarchy_level":-1}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					if hierarchyLevel < 0 {
						return domain.Role{}, errors.New("hierarchy level must be positive")
					}
					return domain.NewRole(name, description, hierarchyLevel, parentRoleID, teamID), nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "zero hierarchy level",
			body: `{"name":"Test","description":"Test role","hierarchy_level":0}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					if hierarchyLevel == 0 {
						return domain.Role{}, errors.New("hierarchy level must be positive")
					}
					return domain.NewRole(name, description, hierarchyLevel, parentRoleID, teamID), nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate role name",
			body: `{"name":"Existing","description":"Test role","hierarchy_level":50}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					return domain.Role{}, errors.New("role name already exists")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "database error",
			body: `{"name":"Test","description":"Test role","hierarchy_level":50}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					return domain.Role{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "very long name",
			body: `{"name":"VeryLongRoleName","description":"Test role","hierarchy_level":50}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.CreateFunc = func(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.Role, error) {
					return domain.Role{}, errors.New("name too long")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "special characters in name",
			body:           `{"name":"Admin<script>","description":"Test role","hierarchy_level":50}`,
			expectedStatus: http.StatusCreated, // Should sanitize
		},
		{
			name:           "with parent role",
			body:           `{"name":"Sub Admin","description":"Test role","hierarchy_level":50,"parent_role_id":"parent-1"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "with team ID",
			body:           `{"name":"Team Admin","description":"Test role","hierarchy_level":50,"team_id":"team-1"}`,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)
			req := httptest.NewRequest("POST", "/roles", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestRolesHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "existing role",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.RoleDetailResponse
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Name != "Admin" {
					t.Errorf("expected name 'Admin', got '%s'", resp.Name)
				}
			},
		},
		{
			name:           "non-existing role",
			roleID:         "non-existent",
			setupMock:      func(m *testutil.MockRolesService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "role with permissions",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.Permissions["role-1"] = []domain.Permission{
					{ID: "perm-1", Code: "rules:read"},
					{ID: "perm-2", Code: "users:write"},
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.RoleDetailResponse
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp.Permissions) != 2 {
					t.Errorf("expected 2 permissions, got %d", len(resp.Permissions))
				}
			},
		},
		{
			name:   "database error",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.GetByIDFunc = func(ctx context.Context, id string) (domain.Role, error) {
					return domain.Role{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)

			r := chi.NewRouter()
			r.Get("/roles/{id}", h.Get)

			req := httptest.NewRequest("GET", "/roles/"+tt.roleID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestRolesHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "list all roles",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.Roles["role-2"] = domain.Role{ID: "role-2", Name: "Member", HierarchyLevel: 50}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.RoleResponse
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 2 {
					t.Errorf("expected 2 roles, got %d", len(resp))
				}
			},
		},
		{
			name:           "empty list",
			setupMock:      func(m *testutil.MockRolesService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "filter by team",
			queryParams: "?team_id=team-1",
			setupMock: func(m *testutil.MockRolesService) {
				teamID := "team-1"
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Team Admin", HierarchyLevel: 100, TeamID: &teamID}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "database error",
			setupMock: func(m *testutil.MockRolesService) {
				m.ListFunc = func(ctx context.Context, teamID *string) ([]domain.Role, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)
			req := httptest.NewRequest("GET", "/roles"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			h.List(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestRolesHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
	}{
		{
			name:   "successful deletion",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "non-existing role",
			roleID:         "non-existent",
			setupMock:      func(m *testutil.MockRolesService) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "system role cannot be deleted",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "System Admin", HierarchyLevel: 100, IsSystem: true}
				m.DeleteFunc = func(ctx context.Context, id string) error {
					return errors.New("cannot modify system role")
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "role has users assigned",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.DeleteFunc = func(ctx context.Context, id string) error {
					return errors.New("role has users assigned")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "database error",
			roleID: "role-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.DeleteFunc = func(ctx context.Context, id string) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)

			r := chi.NewRouter()
			r.Delete("/roles/{id}", h.Delete)

			req := httptest.NewRequest("DELETE", "/roles/"+tt.roleID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRolesHandler_AddPermission(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		body           string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
	}{
		{
			name:   "successful add",
			roleID: "role-1",
			body:   `{"permission_id":"perm-1"}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "role not found",
			roleID:         "non-existent",
			body:           `{"permission_id":"perm-1"}`,
			setupMock:      func(m *testutil.MockRolesService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid JSON",
			roleID: "role-1",
			body:   `{invalid}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "missing permission_id",
			roleID: "role-1",
			body:   `{}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "permission already assigned",
			roleID: "role-1",
			body:   `{"permission_id":"perm-1"}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.Permissions["role-1"] = []domain.Permission{{ID: "perm-1", Code: "rules:read"}}
			},
			expectedStatus: http.StatusNoContent, // Idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)

			r := chi.NewRouter()
			r.Post("/roles/{id}/permissions", h.AddPermission)

			req := httptest.NewRequest("POST", "/roles/"+tt.roleID+"/permissions", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRolesHandler_AssignUser(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		body           string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
	}{
		{
			name:   "successful assignment",
			roleID: "role-1",
			body:   `{"user_id":"user-1"}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "role not found",
			roleID:         "non-existent",
			body:           `{"user_id":"user-1"}`,
			setupMock:      func(m *testutil.MockRolesService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid JSON",
			roleID: "role-1",
			body:   `{invalid}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty user_id field",
			roleID: "role-1",
			body:   `{"user_id":""}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusBadRequest, // Mock validates user_id is not empty
		},
		{
			name:   "user already has role",
			roleID: "role-1",
			body:   `{"user_id":"user-1"}`,
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.UserRoles["user-1"] = []string{"role-1"}
			},
			expectedStatus: http.StatusNoContent, // Idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)

			r := chi.NewRouter()
			r.Post("/roles/{id}/users", h.AssignUser)

			req := httptest.NewRequest("POST", "/roles/"+tt.roleID+"/users", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRolesHandler_RemoveUser(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		userID         string
		setupMock      func(*testutil.MockRolesService)
		expectedStatus int
	}{
		{
			name:   "successful removal",
			roleID: "role-1",
			userID: "user-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
				m.UserRoles["user-1"] = []string{"role-1"}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "role not found for remove user",
			roleID: "non-existent",
			userID: "user-1",
			setupMock: func(m *testutil.MockRolesService) {
				// RemoveUserRole doesn't check role exists first in current implementation
			},
			expectedStatus: http.StatusNoContent, // Current implementation is idempotent
		},
		{
			name:   "user not assigned to role",
			roleID: "role-1",
			userID: "user-1",
			setupMock: func(m *testutil.MockRolesService) {
				m.Roles["role-1"] = domain.Role{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
			},
			expectedStatus: http.StatusNoContent, // Idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRolesService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRolesHandler(svc)

			r := chi.NewRouter()
			r.Delete("/roles/{id}/users/{userId}", h.RemoveUser)

			req := httptest.NewRequest("DELETE", "/roles/"+tt.roleID+"/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRolesHandler_ListPermissions(t *testing.T) {
	svc := testutil.NewMockRolesService()
	h := handlers.NewRolesHandler(svc)

	req := httptest.NewRequest("GET", "/roles/permissions", nil)
	rec := httptest.NewRecorder()

	h.ListPermissions(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []handlers.PermissionResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(resp))
	}
}
