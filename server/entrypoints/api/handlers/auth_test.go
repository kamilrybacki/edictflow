package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/services/auth"
)

type mockAuthService struct {
	registerFunc func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error)
	loginFunc    func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return "test-token", domain.User{ID: "test-user", Email: req.Email, Name: req.Name}, nil
}

func (m *mockAuthService) Login(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return "test-token", domain.User{ID: "test-user", Email: req.Email}, nil
}

type mockUserServiceForAuth struct {
	users map[string]domain.User
}

func (m *mockUserServiceForAuth) GetByID(ctx context.Context, id string) (domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return domain.User{}, errors.New("user not found")
}

func (m *mockUserServiceForAuth) Update(ctx context.Context, user domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserServiceForAuth) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	return nil
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(*mockAuthService)
		expectedStatus int
	}{
		{
			name: "successful registration",
			body: `{"email":"test@example.com","name":"Test User","password":"Password123"}`,
			setupMock: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "jwt-token-123", domain.User{ID: "user-1", Email: req.Email, Name: req.Name}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "registration fails",
			body: `{"email":"test@example.com","name":"Test User","password":"Password123"}`,
			setupMock: func(m *mockAuthService) {
				m.registerFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("email already exists")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := &mockAuthService{}
			if tt.setupMock != nil {
				tt.setupMock(authSvc)
			}
			userSvc := &mockUserServiceForAuth{users: make(map[string]domain.User)}

			h := handlers.NewAuthHandler(authSvc, userSvc)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(*mockAuthService)
		expectedStatus int
	}{
		{
			name: "successful login",
			body: `{"email":"test@example.com","password":"Password123"}`,
			setupMock: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "jwt-token-123", domain.User{ID: "user-1", Email: req.Email}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: `{"email":"test@example.com","password":"wrong"}`,
			setupMock: func(m *mockAuthService) {
				m.loginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("invalid email or password")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := &mockAuthService{}
			if tt.setupMock != nil {
				tt.setupMock(authSvc)
			}
			userSvc := &mockUserServiceForAuth{users: make(map[string]domain.User)}

			h := handlers.NewAuthHandler(authSvc, userSvc)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&resp)
				if _, ok := resp["token"]; !ok {
					t.Error("expected token in response")
				}
			}
		})
	}
}
