// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// This generates a test token for the test infrastructure
	// Secret must match the one in docker-compose.test.yml
	secret := "test-secret-for-local-testing-only"

	claims := jwt.MapClaims{
		"sub":     "c0000000-0000-0000-0000-000000000001", // Test user ID
		"email":   "developer@test.local",
		"name":    "Test Developer",
		"team_id": "a0000000-0000-0000-0000-000000000001",
		"role":    "admin",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	fmt.Println(tokenString)
}
