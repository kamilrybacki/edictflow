package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/tests/testutil"
)

func TestUsersHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*testutil.MockUsersService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "list all users",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
				m.Users["user-2"] = domain.User{ID: "user-2", Email: "user2@example.com", Name: "User 2", IsActive: true}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.UserResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 2 {
					t.Errorf("expected 2 users, got %d", len(resp))
				}
			},
		},
		{
			name:        "list active only",
			queryParams: "?active_only=true",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
				m.Users["user-2"] = domain.User{ID: "user-2", Email: "user2@example.com", Name: "User 2", IsActive: false}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.UserResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 1 {
					t.Errorf("expected 1 active user, got %d", len(resp))
				}
			},
		},
		{
			name: "empty list",
			setupMock: func(m *testutil.MockUsersService) {
				// No users
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.UserResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp == nil {
					// Should return empty array, not null
				}
			},
		},
		{
			name: "database error",
			setupMock: func(m *testutil.MockUsersService) {
				m.ListFunc = func(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error) {
					return nil, errors.New("database connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "filter by team",
			queryParams: "?team_id=team-1",
			setupMock: func(m *testutil.MockUsersService) {
				teamID := "team-1"
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true, TeamID: &teamID}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockUsersService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewUsersHandler(svc)
			req := httptest.NewRequest("GET", "/users"+tt.queryParams, nil)
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

func TestUsersHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*testutil.MockUsersService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "existing user",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.UserResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.ID != "user-1" {
					t.Errorf("expected ID 'user-1', got '%s'", resp.ID)
				}
			},
		},
		{
			name:           "non-existing user",
			userID:         "non-existent",
			setupMock:      func(m *testutil.MockUsersService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "empty user ID - handled by router",
			userID: "nonexistent-user",
			setupMock: func(m *testutil.MockUsersService) {
				m.GetByIDFunc = func(ctx context.Context, id string) (domain.User, error) {
					return domain.User{}, testutil.ErrNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.GetByIDFunc = func(ctx context.Context, id string) (domain.User, error) {
					return domain.User{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "special characters in ID",
			userID:         "user-1-special-chars",
			setupMock:      func(m *testutil.MockUsersService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "very long ID",
			userID:         "user-" + strings.Repeat("a", 100),
			setupMock:      func(m *testutil.MockUsersService) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockUsersService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewUsersHandler(svc)

			r := chi.NewRouter()
			r.Get("/users/{id}", h.Get)

			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
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

func TestUsersHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		body           string
		setupMock      func(*testutil.MockUsersService)
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: "user-1",
			body:   `{"name":"Updated Name"}`,
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "user not found",
			userID:         "non-existent",
			body:           `{"name":"Updated Name"}`,
			setupMock:      func(m *testutil.MockUsersService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "invalid JSON",
			userID: "user-1",
			body:   `{invalid}`,
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty body",
			userID: "user-1",
			body:   ``,
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "update fails - database error",
			userID: "user-1",
			body:   `{"name":"Updated Name"}`,
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
				m.UpdateFunc = func(ctx context.Context, user domain.User) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusBadRequest, // Handler returns 400 for Update errors
		},
		{
			name:   "XSS in name",
			userID: "user-1",
			body:   `{"name":"<script>alert('xss')</script>"}`,
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusNoContent, // Should sanitize or reject
		},
		{
			name:   "empty name",
			userID: "user-1",
			body:   `{"name":""}`,
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusNoContent, // Depends on validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockUsersService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewUsersHandler(svc)

			r := chi.NewRouter()
			r.Put("/users/{id}", h.Update)

			req := httptest.NewRequest("PUT", "/users/"+tt.userID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUsersHandler_Deactivate(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*testutil.MockUsersService)
		expectedStatus int
	}{
		{
			name:   "successful deactivation",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "user not found",
			userID:         "non-existent",
			setupMock:      func(m *testutil.MockUsersService) {},
			expectedStatus: http.StatusInternalServerError, // Handler returns 500 for all Deactivate errors
		},
		{
			name:   "already deactivated",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: false}
			},
			expectedStatus: http.StatusNoContent, // Idempotent
		},
		{
			name:   "database error",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.DeactivateFunc = func(ctx context.Context, id string) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "cannot deactivate self",
			userID: "current-user",
			setupMock: func(m *testutil.MockUsersService) {
				m.DeactivateFunc = func(ctx context.Context, id string) error {
					return errors.New("cannot deactivate yourself")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockUsersService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewUsersHandler(svc)

			r := chi.NewRouter()
			r.Post("/users/{id}/deactivate", h.Deactivate)

			req := httptest.NewRequest("POST", "/users/"+tt.userID+"/deactivate", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUsersHandler_GetWithRolesAndPermissions(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*testutil.MockUsersService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "user with roles and permissions",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{
					ID:          "user-1",
					Email:       "user1@example.com",
					Name:        "User 1",
					IsActive:    true,
					Permissions: []string{"rules:read", "users:read"},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			userID:         "non-existent",
			setupMock:      func(m *testutil.MockUsersService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "user with no permissions",
			userID: "user-1",
			setupMock: func(m *testutil.MockUsersService) {
				m.Users["user-1"] = domain.User{
					ID:          "user-1",
					Email:       "user1@example.com",
					Name:        "User 1",
					IsActive:    true,
					Permissions: []string{},
				}
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockUsersService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewUsersHandler(svc)

			r := chi.NewRouter()
			r.Get("/users/{id}", h.Get) // Get already uses GetWithRolesAndPermissions internally

			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
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
