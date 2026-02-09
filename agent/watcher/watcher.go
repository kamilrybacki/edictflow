// agent/watcher/watcher.go
package watcher

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileInfo struct {
	Path         string
	OriginalHash string
	RuleID       string
}

type ChangeHandler func(path, ruleID, originalHash, newHash, diff string)

type Watcher struct {
	fsWatcher        *fsnotify.Watcher
	files            map[string]FileInfo
	filesMu          sync.RWMutex
	onChangeDetected ChangeHandler
	done             chan struct{}
	pollInterval     time.Duration // If > 0, use polling instead of fsnotify
}

func New() (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		fsWatcher: fsw,
		files:     make(map[string]FileInfo),
		done:      make(chan struct{}),
	}, nil
}

// NewWithPolling creates a watcher that uses periodic file polling
// instead of fsnotify. This is more reliable in container environments.
func NewWithPolling(interval time.Duration) (*Watcher, error) {
	return &Watcher{
		files:        make(map[string]FileInfo),
		done:         make(chan struct{}),
		pollInterval: interval,
	}, nil
}

func (w *Watcher) OnChange(handler ChangeHandler) {
	w.onChangeDetected = handler
}

func (w *Watcher) WatchProject(projectPath, ruleID string) error {
	claudeMDPath := filepath.Join(projectPath, "CLAUDE.md")

	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		// File doesn't exist yet, just watch the directory (fsnotify mode only)
		if w.fsWatcher != nil {
			return w.fsWatcher.Add(projectPath)
		}
		return nil
	}

	hash, err := hashFile(claudeMDPath)
	if err != nil {
		return err
	}

	w.filesMu.Lock()
	w.files[claudeMDPath] = FileInfo{
		Path:         claudeMDPath,
		OriginalHash: hash,
		RuleID:       ruleID,
	}
	w.filesMu.Unlock()

	// In polling mode, we don't need to add to fsWatcher
	if w.fsWatcher != nil {
		return w.fsWatcher.Add(claudeMDPath)
	}
	return nil
}

func (w *Watcher) UnwatchProject(projectPath string) {
	claudeMDPath := filepath.Join(projectPath, "CLAUDE.md")

	if w.fsWatcher != nil {
		_ = w.fsWatcher.Remove(claudeMDPath)
		_ = w.fsWatcher.Remove(projectPath)
	}

	w.filesMu.Lock()
	delete(w.files, claudeMDPath)
	w.filesMu.Unlock()
}

func (w *Watcher) Start() {
	if w.pollInterval > 0 {
		go w.runPolling()
	} else {
		go w.run()
	}
}

func (w *Watcher) Stop() {
	close(w.done)
	if w.fsWatcher != nil {
		w.fsWatcher.Close()
	}
}

func (w *Watcher) run() {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)

		case <-w.done:
			return
		}
	}
}

// runPolling uses periodic file stat checking instead of fsnotify
func (w *Watcher) runPolling() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.pollFiles()
		case <-w.done:
			return
		}
	}
}

// pollFiles checks all watched files for changes.
// Uses batched updates to minimize lock contention.
func (w *Watcher) pollFiles() {
	w.filesMu.RLock()
	files := make(map[string]FileInfo, len(w.files))
	for k, v := range w.files {
		files[k] = v
	}
	w.filesMu.RUnlock()

	// Collect all hash updates to apply in a single lock acquisition
	type hashUpdate struct {
		path    string
		newHash string
	}
	var updates []hashUpdate

	for path, info := range files {
		newHash, err := hashFile(path)
		if err != nil {
			continue
		}

		if newHash != info.OriginalHash {
			if w.onChangeDetected != nil {
				// TODO: Generate actual diff
				w.onChangeDetected(path, info.RuleID, info.OriginalHash, newHash, "")
			}
			updates = append(updates, hashUpdate{path: path, newHash: newHash})
		}
	}

	// Apply all updates in a single lock acquisition
	if len(updates) > 0 {
		w.filesMu.Lock()
		for _, u := range updates {
			if fi, ok := w.files[u.path]; ok {
				fi.OriginalHash = u.newHash
				w.files[u.path] = fi
			}
		}
		w.filesMu.Unlock()
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return
	}

	path := event.Name
	if filepath.Base(path) != "CLAUDE.md" {
		return
	}

	w.filesMu.RLock()
	info, exists := w.files[path]
	w.filesMu.RUnlock()

	if !exists {
		return
	}

	newHash, err := hashFile(path)
	if err != nil {
		return
	}

	if newHash == info.OriginalHash {
		return
	}

	if w.onChangeDetected != nil {
		// TODO: Generate actual diff
		w.onChangeDetected(path, info.RuleID, info.OriginalHash, newHash, "")
	}
}

// hashFile computes SHA256 hash using streaming to avoid loading entire file into memory.
// This is more efficient for large files and reduces memory pressure.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (w *Watcher) RevertFile(path string) error {
	w.filesMu.RLock()
	info, exists := w.files[path]
	w.filesMu.RUnlock()

	if !exists {
		return nil
	}

	// This would need the original content stored somewhere
	_ = info
	return nil
}
