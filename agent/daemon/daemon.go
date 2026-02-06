package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/kamilrybacki/edictflow/agent/notify"
	"github.com/kamilrybacki/edictflow/agent/renderer"
	"github.com/kamilrybacki/edictflow/agent/storage"
	"github.com/kamilrybacki/edictflow/agent/watcher"
	"github.com/kamilrybacki/edictflow/agent/ws"
)

const (
	// Version is the current agent version
	Version = "0.1.0"
	// EnterpriseFilePath is the fixed path for enterprise rules
	EnterpriseFilePath = "/etc/claude-code/CLAUDE.md"
	// UserFileName is the filename for user rules (in ~/.claude/)
	UserFileName = "CLAUDE.md"
	// ProjectFileName is the filename for project rules
	ProjectFileName = "CLAUDE.md"
)

// ManagedFile represents a CLAUDE.md file managed by the daemon
type ManagedFile struct {
	Level string
	Path  string
}

type Daemon struct {
	store        *storage.Storage
	wsClient     *ws.Client
	serverURL    string
	listener     net.Listener
	fileWatcher  *watcher.Watcher
	renderer     *renderer.Renderer
	managedFiles map[string]ManagedFile // path -> level
	projectDirs  []string               // watched project directories
	connectedAt  time.Time              // when the daemon connected
	hostname     string                 // cached hostname
}

func GetPIDFile() (string, error) {
	configDir, err := storage.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "daemon.pid"), nil
}

func GetSocketPath() (string, error) {
	configDir, err := storage.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "daemon.sock"), nil
}

func IsRunning() (int, bool) {
	pidFile, err := GetPIDFile()
	if err != nil {
		return 0, false
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}

	err = process.Signal(syscall.Signal(0))
	return pid, err == nil
}

func Start(serverURL string, foreground bool, pollInterval time.Duration) error {
	if pid, running := IsRunning(); running {
		return fmt.Errorf("daemon already running (PID %d)", pid)
	}

	if foreground {
		return runDaemon(serverURL, pollInterval)
	}

	// Fork child process
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{"start", "--foreground", "--server", serverURL}
	if pollInterval > 0 {
		args = append(args, "--poll-interval", pollInterval.String())
	}

	cmd := exec.Command(executable, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Printf("Daemon started (PID %d)\n", cmd.Process.Pid)
	return nil
}

func Stop() error {
	pid, running := IsRunning()
	if !running {
		return fmt.Errorf("daemon not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	pidFile, _ := GetPIDFile()
	os.Remove(pidFile)

	fmt.Println("Daemon stopped")
	return nil
}

func runDaemon(serverURL string, pollInterval time.Duration) error {
	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer store.Close()

	auth, err := store.GetAuth()
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Write PID file
	pidFile, err := GetPIDFile()
	if err != nil {
		return err
	}
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		log.Printf("Warning: failed to write PID file: %v", err)
	}
	defer os.Remove(pidFile)

	// Create WebSocket client
	wsClient := ws.NewClient(serverURL, auth.AccessToken)

	// Get hostname for heartbeat
	hostname, _ := os.Hostname()

	d := &Daemon{
		store:        store,
		wsClient:     wsClient,
		serverURL:    serverURL,
		renderer:     renderer.New(),
		managedFiles: make(map[string]ManagedFile),
		connectedAt:  time.Now(),
		hostname:     hostname,
	}

	// Initialize managed file paths
	d.initManagedFiles()

	// Create file watcher with optional polling mode
	var fw *watcher.Watcher
	if pollInterval > 0 {
		fw, err = watcher.NewWithPolling(pollInterval)
		log.Printf("Using polling file watcher with interval %v", pollInterval)
	} else {
		fw, err = watcher.New()
	}
	if err != nil {
		log.Printf("Failed to create file watcher: %v", err)
	} else {
		d.fileWatcher = fw
		d.setupFileWatcher()
		fw.Start()
		defer fw.Stop()
	}

	// Setup handlers
	d.setupHandlers()

	// Start Unix socket for CLI queries
	if err := d.startSocket(); err != nil {
		log.Printf("Failed to start socket: %v", err)
	}
	defer d.stopSocket()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Connect to server
	go wsClient.ConnectWithRetry()

	log.Println("Daemon running...")
	<-sigChan
	log.Println("Shutting down...")

	wsClient.Close()
	return nil
}

func (d *Daemon) setupHandlers() {
	d.wsClient.OnConnect(func() {
		log.Println("Connected to server")
		notify.ConnectionRestored()
		d.sendHeartbeat()
	})

	d.wsClient.OnDisconnect(func() {
		log.Println("Disconnected from server")
		notify.ConnectionLost()
	})

	d.wsClient.OnMessage(ws.TypeConfigUpdate, d.handleConfigUpdate)
	d.wsClient.OnMessage(ws.TypeAck, d.handleAck)
	d.wsClient.OnMessage(ws.TypeChangeApproved, d.handleChangeApproved)
	d.wsClient.OnMessage(ws.TypeChangeRejected, d.handleChangeRejected)
}

func (d *Daemon) sendHeartbeat() {
	projects, _ := d.store.GetProjects()
	paths := make([]string, len(projects))
	for i, p := range projects {
		paths[i] = p.Path
	}

	payload := ws.HeartbeatPayload{
		Status:         "online",
		CachedVersion:  d.store.GetCachedVersion(),
		ActiveProjects: paths,
		Hostname:       d.hostname,
		Version:        Version,
		OS:             runtime.GOOS + "/" + runtime.GOARCH,
		ConnectedAt:    d.connectedAt.Format(time.RFC3339),
	}

	msg, _ := ws.NewMessage(ws.TypeHeartbeat, payload)
	d.wsClient.Send(msg)
}

func (d *Daemon) handleConfigUpdate(msg ws.Message) {
	var payload ws.ConfigUpdatePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Invalid config update: %v", err)
		return
	}

	rules := make([]storage.CachedRule, len(payload.Rules))
	for i, r := range payload.Rules {
		rules[i] = storage.CachedRule{
			ID:                    r.ID,
			Name:                  r.Name,
			Content:               r.Content,
			TargetLayer:           r.TargetLayer,
			Triggers:              r.Triggers,
			EnforcementMode:       r.EnforcementMode,
			TemporaryTimeoutHours: r.TemporaryTimeoutHours,
			Version:               payload.Version,
		}
	}

	if err := d.store.SaveRules(rules, payload.Version); err != nil {
		log.Printf("Failed to save rules: %v", err)
	} else {
		log.Printf("Updated rules to version %d", payload.Version)
		notify.ConfigUpdated(payload.Version)
	}
}

func (d *Daemon) handleAck(msg ws.Message) {
	var payload ws.AckPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}
	d.store.DeleteMessage(payload.RefID)
}

func (d *Daemon) handleChangeApproved(msg ws.Message) {
	var payload ws.ChangeApprovedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}
	log.Printf("Change %s approved", payload.ChangeID)
	d.store.UpdateChangeStatus(payload.ChangeID, "approved")
	notify.ChangeApproved(payload.ChangeID)
}

func (d *Daemon) handleChangeRejected(msg ws.Message) {
	var payload ws.ChangeRejectedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}
	log.Printf("Change %s rejected", payload.ChangeID)
	d.store.UpdateChangeStatus(payload.ChangeID, "rejected")
	notify.ChangeRejected(payload.ChangeID)
	// TODO: Revert file to original content
}

func (d *Daemon) setupFileWatcher() {
	d.fileWatcher.OnChange(func(path, ruleID, originalHash, newHash, diff string) {
		log.Printf("Change detected in %s", path)
		notify.ChangeBlocked(path)

		// Send change_detected message to server
		payload := ws.ChangeDetectedPayload{
			RuleID:       ruleID,
			FilePath:     path,
			OriginalHash: originalHash,
			ModifiedHash: newHash,
			Diff:         diff,
		}
		msg, _ := ws.NewMessage(ws.TypeChangeDetected, payload)
		d.wsClient.Send(msg)
	})

	// Watch all projects from storage
	projects, _ := d.store.GetProjects()
	for _, p := range projects {
		if err := d.fileWatcher.WatchProject(p.Path, ""); err != nil {
			log.Printf("Failed to watch project %s: %v", p.Path, err)
		}
	}
}

// initManagedFiles sets up the three fixed-location CLAUDE.md files
func (d *Daemon) initManagedFiles() {
	// Enterprise file (system-wide)
	d.managedFiles[EnterpriseFilePath] = ManagedFile{
		Level: "enterprise",
		Path:  EnterpriseFilePath,
	}

	// User file (~/.claude/CLAUDE.md)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userPath := filepath.Join(homeDir, ".claude", UserFileName)
		d.managedFiles[userPath] = ManagedFile{
			Level: "user",
			Path:  userPath,
		}
	}

	// Project files are added dynamically when watching directories
}

// getUserFilePath returns the path to the user's CLAUDE.md file
func getUserFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude", UserFileName), nil
}

// AddProjectDirectory registers a project directory for watching
func (d *Daemon) AddProjectDirectory(projectPath string) error {
	claudePath := filepath.Join(projectPath, ProjectFileName)
	d.managedFiles[claudePath] = ManagedFile{
		Level: "project",
		Path:  claudePath,
	}
	d.projectDirs = append(d.projectDirs, projectPath)

	// Sync the file immediately
	return d.syncFile("project", claudePath)
}

// syncFile renders and writes the managed section for a specific level/path
func (d *Daemon) syncFile(level string, path string) error {
	rules, err := d.store.GetRulesByLayer(level)
	if err != nil {
		return fmt.Errorf("failed to get rules for %s: %w", level, err)
	}

	managed := d.renderer.RenderManagedSection(rules)

	// Read existing content (if any)
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Merge with existing content
	merged := d.renderer.MergeWithFile(string(existing), managed)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(path, []byte(merged), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	log.Printf("Synced %s CLAUDE.md at %s", level, path)
	return nil
}

// SyncAllFiles syncs all managed CLAUDE.md files
func (d *Daemon) SyncAllFiles() error {
	for path, file := range d.managedFiles {
		if err := d.syncFile(file.Level, path); err != nil {
			// Log but continue - enterprise path might not be writable
			log.Printf("Failed to sync %s: %v", path, err)
		}
	}
	return nil
}

// CheckAndRestoreTamperedFiles checks if any managed sections were modified and restores them
func (d *Daemon) CheckAndRestoreTamperedFiles() {
	for path, file := range d.managedFiles {
		rules, err := d.store.GetRulesByLayer(file.Level)
		if err != nil {
			continue
		}

		expected := d.renderer.RenderManagedSection(rules)

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if d.renderer.DetectManagedSectionTampering(string(content), expected) {
			log.Printf("Tampering detected in %s, restoring...", path)
			if err := d.syncFile(file.Level, path); err != nil {
				log.Printf("Failed to restore %s: %v", path, err)
			} else {
				notify.ManagedSectionRestored(path)
			}
		}
	}
}
