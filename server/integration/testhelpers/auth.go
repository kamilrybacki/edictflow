//go:build integration

package testhelpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TestJWTSecret = "test-integration-secret"

// Claims represents the JWT claims for test tokens
type Claims struct {
	jwt.RegisteredClaims
	TeamID string `json:"team_id,omitempty"`
}

// GenerateTestToken creates a JWT token for testing
func GenerateTestToken(userID, teamID string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		TeamID: teamID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// AuthHeader returns the Authorization header value for a test token
func AuthHeader(userID, teamID string) (string, error) {
	token, err := GenerateTestToken(userID, teamID)
	if err != nil {
		return "", err
	}
	return "Bearer " + token, nil
}
