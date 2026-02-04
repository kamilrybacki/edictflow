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
	"strconv"
	"syscall"

	"github.com/kamilrybacki/claudeception/agent/notify"
	"github.com/kamilrybacki/claudeception/agent/storage"
	"github.com/kamilrybacki/claudeception/agent/watcher"
	"github.com/kamilrybacki/claudeception/agent/ws"
)

type Daemon struct {
	store       *storage.Storage
	wsClient    *ws.Client
	serverURL   string
	listener    net.Listener
	fileWatcher *watcher.Watcher
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

func Start(serverURL string, foreground bool) error {
	if pid, running := IsRunning(); running {
		return fmt.Errorf("daemon already running (PID %d)", pid)
	}

	if foreground {
		return runDaemon(serverURL)
	}

	// Fork child process
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, "start", "--foreground", "--server", serverURL)
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

func runDaemon(serverURL string) error {
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
	os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
	defer os.Remove(pidFile)

	// Create WebSocket client
	wsClient := ws.NewClient(serverURL, auth.AccessToken)

	d := &Daemon{
		store:     store,
		wsClient:  wsClient,
		serverURL: serverURL,
	}

	// Create file watcher
	fw, err := watcher.New()
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
