package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/services/auth"
	"github.com/kamilrybacki/edictflow/server/tests/testutil"
)

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(*testutil.MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			body: `{"email":"test@example.com","name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "jwt-token-123", domain.User{ID: "user-1", Email: req.Email, Name: req.Name}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if _, ok := resp["token"]; !ok {
					t.Error("expected token in response")
				}
			},
		},
		{
			name:           "invalid JSON body",
			body:           `{"email": invalid}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			body: `{"name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					if req.Email == "" {
						return "", domain.User{}, errors.New("email is required")
					}
					return "test-token", domain.User{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			body: `{"email":"test@example.com","name":"Test User"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					if req.Password == "" {
						return "", domain.User{}, errors.New("password is required")
					}
					return "test-token", domain.User{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing name",
			body: `{"email":"test@example.com","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					if req.Name == "" {
						return "", domain.User{}, errors.New("name is required")
					}
					return "test-token", domain.User{}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "email already exists",
			body: `{"email":"existing@example.com","name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("email already exists")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "weak password",
			body: `{"email":"test@example.com","name":"Test User","password":"weak"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("password does not meet requirements")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "database error",
			body: `{"email":"test@example.com","name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("database connection failed")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			body: `{"email":"not-an-email","name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("invalid email format")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "very long email",
			body: `{"email":"` + strings.Repeat("a", 500) + `@example.com","name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("email too long")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "SQL injection attempt in email",
			body: `{"email":"test@example.com'; DROP TABLE users;--","name":"Test User","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("invalid email format")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "XSS attempt in name",
			body: `{"email":"test@example.com","name":"<script>alert('xss')</script>","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("invalid name format")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := &testutil.MockAuthService{}
			if tt.setupMock != nil {
				tt.setupMock(authSvc)
			}
			userSvc := testutil.NewMockUserServiceForAuth()

			h := handlers.NewAuthHandler(authSvc, userSvc)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(*testutil.MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			body: `{"email":"test@example.com","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "jwt-token-123", domain.User{ID: "user-1", Email: req.Email, Name: "Test User"}, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				_ = json.NewDecoder(rec.Body).Decode(&resp)
				if _, ok := resp["token"]; !ok {
					t.Error("expected token in response")
				}
			},
		},
		{
			name:           "invalid JSON body",
			body:           `not json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			body: `{"password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					if req.Email == "" {
						return "", domain.User{}, errors.New("email is required")
					}
					return "test-token", domain.User{}, nil
				}
			},
			expectedStatus: http.StatusUnauthorized, // Handler returns 401 for any login error
		},
		{
			name: "missing password",
			body: `{"email":"test@example.com"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					if req.Password == "" {
						return "", domain.User{}, errors.New("password is required")
					}
					return "test-token", domain.User{}, nil
				}
			},
			expectedStatus: http.StatusUnauthorized, // Handler returns 401 for any login error
		},
		{
			name: "invalid credentials - wrong password",
			body: `{"email":"test@example.com","password":"wrong"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("invalid email or password")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid credentials - user not found",
			body: `{"email":"nonexistent@example.com","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("invalid email or password")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "account locked",
			body: `{"email":"locked@example.com","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("account locked")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "account deactivated",
			body: `{"email":"deactivated@example.com","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("account deactivated")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "database error",
			body: `{"email":"test@example.com","password":"Password123"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("database connection failed")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "extremely long password",
			body: `{"email":"test@example.com","password":"` + strings.Repeat("a", 10000) + `"}`,
			setupMock: func(m *testutil.MockAuthService) {
				m.LoginFunc = func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
					return "", domain.User{}, errors.New("password too long")
				}
			},
			expectedStatus: http.StatusUnauthorized, // Handler returns 401 for any login error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := &testutil.MockAuthService{}
			if tt.setupMock != nil {
				tt.setupMock(authSvc)
			}
			userSvc := testutil.NewMockUserServiceForAuth()

			h := handlers.NewAuthHandler(authSvc, userSvc)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestAuthHandler_ContentType(t *testing.T) {
	authSvc := &testutil.MockAuthService{}
	userSvc := testutil.NewMockUserServiceForAuth()
	h := handlers.NewAuthHandler(authSvc, userSvc)

	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{"no content type", "", `{"email":"test@example.com","password":"Password123"}`},
		{"wrong content type", "text/plain", `{"email":"test@example.com","password":"Password123"}`},
		{"form urlencoded", "application/x-www-form-urlencoded", "email=test@example.com&password=Password123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			// Should handle gracefully (either parse or return bad request)
			if rec.Code != http.StatusOK && rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnauthorized {
				t.Errorf("unexpected status %d", rec.Code)
			}
		})
	}
}

func TestAuthHandler_LargePayload(t *testing.T) {
	authSvc := &testutil.MockAuthService{
		RegisterFunc: func(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error) {
			return "", domain.User{}, errors.New("payload too large")
		},
	}
	userSvc := testutil.NewMockUserServiceForAuth()
	h := handlers.NewAuthHandler(authSvc, userSvc)

	// Create a large JSON payload
	largePayload := `{"email":"test@example.com","password":"Password123","extra":"` + strings.Repeat("x", 1<<20) + `"}`

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(largePayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	// Should return error status when service rejects
	if rec.Code == http.StatusOK || rec.Code == http.StatusCreated {
		t.Error("expected error for large payload")
	}
}

func TestAuthHandler_ConcurrentRequests(t *testing.T) {
	authSvc := &testutil.MockAuthService{
		LoginFunc: func(ctx context.Context, req auth.LoginRequest) (string, domain.User, error) {
			return "token", domain.User{ID: "user-1", Email: req.Email}, nil
		},
	}
	userSvc := testutil.NewMockUserServiceForAuth()
	h := handlers.NewAuthHandler(authSvc, userSvc)

	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(`{"email":"test@example.com","password":"Password123"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Login(rec, req)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestAuthHandler_RequestBodyClosed(t *testing.T) {
	authSvc := &testutil.MockAuthService{}
	userSvc := testutil.NewMockUserServiceForAuth()
	h := handlers.NewAuthHandler(authSvc, userSvc)

	body := io.NopCloser(bytes.NewBufferString(`{"email":"test@example.com","password":"Password123"}`))
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	// Should not panic when body is already processed
}
