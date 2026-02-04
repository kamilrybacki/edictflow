//go:build tools
// +build tools

// Package tools imports dependencies used by the agent module.
// This file ensures go mod tidy keeps these dependencies in go.mod
// until they are used in actual code.
package tools

import (
	_ "github.com/fsnotify/fsnotify"
	_ "github.com/gen2brain/beeep"
	_ "github.com/google/uuid"
	_ "github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/sergi/go-diff/diffmatchpatch"
)
