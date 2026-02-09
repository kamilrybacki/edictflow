package daemon

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
)

type StatusResponse struct {
	Running       bool     `json:"running"`
	Connected     bool     `json:"connected"`
	CachedVersion int      `json:"cached_version"`
	Projects      []string `json:"projects"`
	PendingMsgs   int      `json:"pending_messages"`
}

func (d *Daemon) startSocket() error {
	socketPath, err := GetSocketPath()
	if err != nil {
		return err
	}

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	d.listener = listener

	go d.acceptConnections()
	return nil
}

func (d *Daemon) stopSocket() {
	if d.listener != nil {
		d.listener.Close()
	}
	socketPath, _ := GetSocketPath()
	os.Remove(socketPath)
}

func (d *Daemon) acceptConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			return
		}
		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	var request map[string]string
	if err := json.Unmarshal([]byte(line), &request); err != nil {
		return
	}

	switch request["command"] {
	case "status":
		d.handleStatusRequest(conn)
	case "sync":
		d.handleSyncRequest(conn)
	}
}

func (d *Daemon) handleStatusRequest(conn net.Conn) {
	projects, _ := d.store.GetProjects()
	paths := make([]string, len(projects))
	for i, p := range projects {
		paths[i] = p.Path
	}

	pending, _ := d.store.GetPendingMessages()

	response := StatusResponse{
		Running:       true,
		Connected:     d.wsClient.State() == 2, // StateConnected
		CachedVersion: d.store.GetCachedVersion(),
		Projects:      paths,
		PendingMsgs:   len(pending),
	}

	data, _ := json.Marshal(response)
	_, _ = conn.Write(append(data, '\n'))
}

func (d *Daemon) handleSyncRequest(conn net.Conn) {
	d.sendHeartbeat()
	_, _ = conn.Write([]byte(`{"status":"sync_requested"}` + "\n"))
}

func QueryDaemon(command string) ([]byte, error) {
	socketPath, err := GetSocketPath()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := map[string]string{"command": command}
	data, _ := json.Marshal(request)
	_, _ = conn.Write(append(data, '\n'))

	reader := bufio.NewReader(conn)
	response, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	return response, nil
}
