// agent/watcher/watcher.go
package watcher

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FileInfo struct {
	Path        string
	OriginalHash string
	RuleID      string
}

type ChangeHandler func(path, ruleID, originalHash, newHash, diff string)

type Watcher struct {
	fsWatcher     *fsnotify.Watcher
	files         map[string]FileInfo
	filesMu       sync.RWMutex
	onChangeDetected ChangeHandler
	done          chan struct{}
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

func (w *Watcher) OnChange(handler ChangeHandler) {
	w.onChangeDetected = handler
}

func (w *Watcher) WatchProject(projectPath, ruleID string) error {
	claudeMDPath := filepath.Join(projectPath, "CLAUDE.md")

	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		// File doesn't exist yet, just watch the directory
		return w.fsWatcher.Add(projectPath)
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

	return w.fsWatcher.Add(claudeMDPath)
}

func (w *Watcher) UnwatchProject(projectPath string) {
	claudeMDPath := filepath.Join(projectPath, "CLAUDE.md")
	w.fsWatcher.Remove(claudeMDPath)
	w.fsWatcher.Remove(projectPath)

	w.filesMu.Lock()
	delete(w.files, claudeMDPath)
	w.filesMu.Unlock()
}

func (w *Watcher) Start() {
	go w.run()
}

func (w *Watcher) Stop() {
	close(w.done)
	w.fsWatcher.Close()
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

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
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
