package deviceauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrExpired              = errors.New("device code expired")
	ErrAuthorizationPending = errors.New("authorization_pending")
	ErrNotFound             = errors.New("device code not found")
)

type TokenGenerator interface {
	GenerateToken(userID string) (string, error)
}

type Service struct {
	repo         Repository
	tokenGen     TokenGenerator
	expiresIn    time.Duration
	pollInterval time.Duration
}

func NewService(repo Repository, tokenGen TokenGenerator) *Service {
	return &Service{
		repo:         repo,
		tokenGen:     tokenGen,
		expiresIn:    15 * time.Minute,
		pollInterval: 5 * time.Second,
	}
}

type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *Service) InitiateDeviceAuth(ctx context.Context, clientID, baseURL string) (DeviceAuthResponse, error) {
	deviceCode := uuid.New().String()
	userCode := generateUserCode()

	dc := domain.DeviceCode{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ClientID:   clientID,
		ExpiresAt:  time.Now().Add(s.expiresIn),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, dc); err != nil {
		return DeviceAuthResponse{}, err
	}

	return DeviceAuthResponse{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: fmt.Sprintf("%s/auth/device/verify", baseURL),
		ExpiresIn:       int(s.expiresIn.Seconds()),
		Interval:        int(s.pollInterval.Seconds()),
	}, nil
}

func (s *Service) PollForToken(ctx context.Context, deviceCode string) (TokenResponse, error) {
	dc, err := s.repo.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		return TokenResponse{}, ErrNotFound
	}

	if dc.IsExpired() {
		_ = s.repo.Delete(ctx, deviceCode)
		return TokenResponse{}, ErrExpired
	}

	if !dc.IsAuthorized() {
		return TokenResponse{}, ErrAuthorizationPending
	}

	token, err := s.tokenGen.GenerateToken(*dc.UserID)
	if err != nil {
		return TokenResponse{}, err
	}

	_ = s.repo.Delete(ctx, deviceCode)

	return TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   86400,
	}, nil
}

func (s *Service) AuthorizeDevice(ctx context.Context, userCode, userID string) error {
	dc, err := s.repo.GetByUserCode(ctx, userCode)
	if err != nil {
		return ErrNotFound
	}

	if dc.IsExpired() {
		_ = s.repo.Delete(ctx, dc.DeviceCode)
		return ErrExpired
	}

	return s.repo.Authorize(ctx, dc.DeviceCode, userID)
}

func (s *Service) GetByUserCode(ctx context.Context, userCode string) (domain.DeviceCode, error) {
	return s.repo.GetByUserCode(ctx, userCode)
}

func generateUserCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	code := strings.ToUpper(base64.StdEncoding.EncodeToString(b))
	code = strings.ReplaceAll(code, "+", "X")
	code = strings.ReplaceAll(code, "/", "Y")
	code = strings.ReplaceAll(code, "=", "")
	if len(code) > 8 {
		code = code[:8]
	}
	return code[:4] + "-" + code[4:]
}
