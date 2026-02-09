// server/entrypoints/api/handlers/device_auth.go
package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/deviceauth"
)

type DeviceAuthService interface {
	InitiateDeviceAuth(ctx context.Context, clientID, baseURL string) (deviceauth.DeviceAuthResponse, error)
	PollForToken(ctx context.Context, deviceCode string) (deviceauth.TokenResponse, error)
	AuthorizeDevice(ctx context.Context, userCode, userID string) error
	GetByUserCode(ctx context.Context, userCode string) (domain.DeviceCode, error)
}

type DeviceAuthHandler struct {
	service DeviceAuthService
	baseURL string
}

func NewDeviceAuthHandler(service DeviceAuthService, baseURL string) *DeviceAuthHandler {
	return &DeviceAuthHandler{
		service: service,
		baseURL: baseURL,
	}
}

type DeviceCodeRequest struct {
	ClientID string `json:"client_id"`
}

type DeviceTokenRequest struct {
	DeviceCode string `json:"device_code"`
	ClientID   string `json:"client_id"`
}

type AuthorizeRequest struct {
	UserCode string `json:"user_code"`
}

func (h *DeviceAuthHandler) InitiateDeviceAuth(w http.ResponseWriter, r *http.Request) {
	var req DeviceCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.ClientID = "edictflow-cli"
	}

	resp, err := h.service.InitiateDeviceAuth(r.Context(), req.ClientID, h.baseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *DeviceAuthHandler) PollForToken(w http.ResponseWriter, r *http.Request) {
	var req DeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := h.service.PollForToken(r.Context(), req.DeviceCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch err {
		case deviceauth.ErrAuthorizationPending:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
		case deviceauth.ErrExpired:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "expired_token"})
		case deviceauth.ErrNotFound:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "server_error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

var verifyTemplate = template.Must(template.New("verify").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authorize Device - Edictflow</title>
    <style>
        body { font-family: system-ui; max-width: 400px; margin: 50px auto; padding: 20px; }
        .code { font-size: 2em; letter-spacing: 0.2em; text-align: center; padding: 20px; background: #f0f0f0; border-radius: 8px; }
        button { width: 100%; padding: 12px; font-size: 1.1em; margin-top: 20px; cursor: pointer; }
        .success { color: green; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>Authorize Device</h1>
    {{if .Error}}
        <p class="error">{{.Error}}</p>
    {{else if .Success}}
        <p class="success">Device authorized! You can close this window.</p>
    {{else}}
        <p>Enter the code shown on your device:</p>
        <form method="POST">
            <input type="text" name="user_code" placeholder="ABCD-1234" style="width:100%;padding:12px;font-size:1.2em;text-align:center;">
            <button type="submit">Authorize</button>
        </form>
    {{end}}
</body>
</html>
`))

func (h *DeviceAuthHandler) VerifyPage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Redirect(w, r, "/login?redirect=/auth/device/verify", http.StatusFound)
		return
	}

	data := struct {
		Error   string
		Success bool
	}{}

	if r.Method == http.MethodPost {
		userCode := r.FormValue("user_code")
		if userCode == "" {
			data.Error = "Please enter a code"
		} else {
			err := h.service.AuthorizeDevice(r.Context(), userCode, userID)
			if err != nil {
				switch err {
				case deviceauth.ErrNotFound:
					data.Error = "Invalid code"
				case deviceauth.ErrExpired:
					data.Error = "Code expired"
				default:
					data.Error = "Authorization failed"
				}
			} else {
				data.Success = true
			}
		}
	}

	code := r.URL.Query().Get("code")
	if code != "" && r.Method == http.MethodGet {
		_, err := h.service.GetByUserCode(r.Context(), code)
		if err == nil {
			err = h.service.AuthorizeDevice(r.Context(), code, userID)
			if err == nil {
				data.Success = true
			}
		}
	}

	w.Header().Set("Content-Type", "text/html")
	_ = verifyTemplate.Execute(w, data)
}

func (h *DeviceAuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/device", h.InitiateDeviceAuth)
	r.Post("/device/token", h.PollForToken)
	r.Get("/device/verify", h.VerifyPage)
	r.Post("/device/verify", h.VerifyPage)
}
