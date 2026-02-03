package db_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/claudeception/server/common/db"
)

func TestNewPoolReturnsErrorForInvalidURL(t *testing.T) {
	_, err := db.NewPool(context.Background(), "invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
