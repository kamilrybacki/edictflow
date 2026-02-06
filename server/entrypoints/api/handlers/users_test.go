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

type mockUsersService struct {
	users map[string]domain.User
}

func newMockUsersService() *mockUsersService {
	return &mockUsersService{users: make(map[string]domain.User)}
}

func (m *mockUsersService) GetByID(ctx context.Context, id string) (domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return domain.User{}, errors.New("user not found")
}

func (m *mockUsersService) List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error) {
	var result []domain.User
	for _, u := range m.users {
		if activeOnly && !u.IsActive {
			continue
		}
		result = append(result, u)
	}
	return result, nil
}

func (m *mockUsersService) Update(ctx context.Context, user domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUsersService) Deactivate(ctx context.Context, id string) error {
	if user, ok := m.users[id]; ok {
		user.IsActive = false
		m.users[id] = user
		return nil
	}
	return errors.New("user not found")
}

func (m *mockUsersService) GetWithRolesAndPermissions(ctx context.Context, id string) (domain.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUsersService) LeaveTeam(ctx context.Context, userID string) error {
	if user, ok := m.users[userID]; ok {
		user.TeamID = nil
		m.users[userID] = user
		return nil
	}
	return errors.New("user not found")
}

func TestUsersHandler_List(t *testing.T) {
	svc := newMockUsersService()
	svc.users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}
	svc.users["user-2"] = domain.User{ID: "user-2", Email: "user2@example.com", Name: "User 2", IsActive: true}

	h := handlers.NewUsersHandler(svc)
	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []handlers.UserResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp))
	}
}

func TestUsersHandler_Get(t *testing.T) {
	svc := newMockUsersService()
	svc.users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}

	h := handlers.NewUsersHandler(svc)

	t.Run("existing user", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users/{id}", h.Get)

		req := httptest.NewRequest("GET", "/users/user-1", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp handlers.UserResponse
		json.NewDecoder(rec.Body).Decode(&resp)

		if resp.ID != "user-1" {
			t.Errorf("expected ID 'user-1', got '%s'", resp.ID)
		}
	})

	t.Run("non-existing user", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/users/{id}", h.Get)

		req := httptest.NewRequest("GET", "/users/non-existent", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestUsersHandler_Deactivate(t *testing.T) {
	svc := newMockUsersService()
	svc.users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}

	h := handlers.NewUsersHandler(svc)

	r := chi.NewRouter()
	r.Post("/users/{id}/deactivate", h.Deactivate)

	req := httptest.NewRequest("POST", "/users/user-1/deactivate", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if svc.users["user-1"].IsActive {
		t.Error("expected user to be deactivated")
	}
}

func TestUsersHandler_Update(t *testing.T) {
	svc := newMockUsersService()
	svc.users["user-1"] = domain.User{ID: "user-1", Email: "user1@example.com", Name: "User 1", IsActive: true}

	h := handlers.NewUsersHandler(svc)

	r := chi.NewRouter()
	r.Put("/users/{id}", h.Update)

	body := `{"name":"Updated Name"}`
	req := httptest.NewRequest("PUT", "/users/user-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}
