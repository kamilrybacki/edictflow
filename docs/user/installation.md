# Agent Installation

Install the Edictflow agent on your development machine.

## System Requirements

| Requirement | Details |
|-------------|---------|
| OS | Linux, macOS, Windows |
| Architecture | amd64, arm64 |
| RAM | 50 MB |
| Disk | 20 MB |
| Network | HTTPS/WSS to server |

## Download

### From GitHub Releases

Download the latest release for your platform:

=== "macOS (Intel)"

    ```bash
    curl -fsSL https://github.com/kamilrybacki/edictflow/releases/latest/download/agent-darwin-amd64 \
      -o edictflow-agent
    chmod +x edictflow-agent
    ```

=== "macOS (Apple Silicon)"

    ```bash
    curl -fsSL https://github.com/kamilrybacki/edictflow/releases/latest/download/agent-darwin-arm64 \
      -o edictflow-agent
    chmod +x edictflow-agent
    ```

=== "Linux (x64)"

    ```bash
    curl -fsSL https://github.com/kamilrybacki/edictflow/releases/latest/download/agent-linux-amd64 \
      -o edictflow-agent
    chmod +x edictflow-agent
    ```

=== "Linux (ARM64)"

    ```bash
    curl -fsSL https://github.com/kamilrybacki/edictflow/releases/latest/download/agent-linux-arm64 \
      -o edictflow-agent
    chmod +x edictflow-agent
    ```

=== "Windows"

    ```powershell
    Invoke-WebRequest `
      -Uri "https://github.com/kamilrybacki/edictflow/releases/latest/download/agent-windows-amd64.exe" `
      -OutFile "edictflow-agent.exe"
    ```

### From Package Manager

=== "Homebrew (macOS/Linux)"

    ```bash
    brew install kamilrybacki/tap/edictflow-agent
    ```

=== "APT (Debian/Ubuntu)"

    ```bash
    curl -fsSL https://edictflow.dev/apt/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/edictflow.gpg
    echo "deb [signed-by=/usr/share/keyrings/edictflow.gpg] https://edictflow.dev/apt stable main" | sudo tee /etc/apt/sources.list.d/edictflow.list
    sudo apt update
    sudo apt install edictflow-agent
    ```

=== "Scoop (Windows)"

    ```powershell
    scoop bucket add edictflow https://github.com/kamilrybacki/scoop-bucket
    scoop install edictflow-agent
    ```

### Build from Source

```bash
# Clone repository
git clone https://github.com/kamilrybacki/edictflow.git
cd edictflow/agent

# Build
go build -o edictflow-agent ./cmd/agent

# Install
sudo mv edictflow-agent /usr/local/bin/
```

## Installation Path

### macOS/Linux

Recommended: `/usr/local/bin/`

```bash
sudo mv edictflow-agent /usr/local/bin/
```

Or user-local: `~/.local/bin/`

```bash
mkdir -p ~/.local/bin
mv edictflow-agent ~/.local/bin/

# Add to PATH if needed
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Windows

Add to a directory in your PATH, or:

```powershell
# Create directory
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\Edictflow"

# Move binary
Move-Item edictflow-agent.exe "$env:LOCALAPPDATA\Edictflow\"

# Add to PATH
$env:Path += ";$env:LOCALAPPDATA\Edictflow"
[Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::User)
```

## Verify Installation

```bash
edictflow-agent version
```

Expected output:

```
edictflow-agent version 1.0.0
  Built: 2024-01-15T10:30:00Z
  Commit: abc1234
```

## Configuration

### Data Directory

The agent stores configuration and cache in:

| Platform | Path |
|----------|------|
| macOS | `~/Library/Application Support/edictflow/` |
| Linux | `~/.config/edictflow/` |
| Windows | `%APPDATA%\edictflow\` |

### Override Data Directory

```bash
export EDICTFLOW_DATA_DIR=/custom/path
```

## Initial Setup

### 1. Login to Server

Authenticate with your organization's Edictflow server:

```bash
edictflow-agent login https://edictflow.yourcompany.com
```

This initiates the device code flow:

1. Agent displays a code
2. Opens browser to verification URL
3. You enter the code and authenticate
4. Agent receives credentials

### 2. Verify Connection

```bash
edictflow-agent status
```

### 3. Start the Agent

```bash
# Start as background daemon
edictflow-agent start

# Or run in foreground for debugging
edictflow-agent start --foreground
```

## Run as Service

### macOS (launchd)

Create `~/Library/LaunchAgents/com.edictflow.agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.edictflow.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/edictflow-agent</string>
        <string>start</string>
        <string>--foreground</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/edictflow-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/edictflow-agent.log</string>
</dict>
</plist>
```

Load the service:

```bash
launchctl load ~/Library/LaunchAgents/com.edictflow.agent.plist
```

### Linux (systemd)

Create `~/.config/systemd/user/edictflow-agent.service`:

```ini
[Unit]
Description=Edictflow Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/edictflow-agent start --foreground
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
```

Enable and start:

```bash
systemctl --user daemon-reload
systemctl --user enable edictflow-agent
systemctl --user start edictflow-agent
```

### Windows (Task Scheduler)

```powershell
$action = New-ScheduledTaskAction -Execute "edictflow-agent.exe" -Argument "start --foreground"
$trigger = New-ScheduledTaskTrigger -AtLogon
$principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -RunLevel Limited
Register-ScheduledTask -TaskName "Edictflow Agent" -Action $action -Trigger $trigger -Principal $principal
```

## Update

### Manual Update

Download and replace the binary:

```bash
curl -fsSL https://github.com/kamilrybacki/edictflow/releases/latest/download/agent-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) \
  -o /tmp/edictflow-agent
chmod +x /tmp/edictflow-agent
sudo mv /tmp/edictflow-agent /usr/local/bin/edictflow-agent
```

Restart the agent:

```bash
edictflow-agent stop
edictflow-agent start
```

### Self-Update

```bash
edictflow-agent update
```

This downloads and installs the latest version.

## Uninstall

### Remove Binary

```bash
# macOS/Linux
sudo rm /usr/local/bin/edictflow-agent

# Windows
Remove-Item "$env:LOCALAPPDATA\Edictflow\edictflow-agent.exe"
```

### Remove Data

```bash
# macOS
rm -rf ~/Library/Application\ Support/edictflow/

# Linux
rm -rf ~/.config/edictflow/

# Windows
Remove-Item -Recurse "$env:APPDATA\edictflow"
```

### Remove Service

```bash
# macOS
launchctl unload ~/Library/LaunchAgents/com.edictflow.agent.plist
rm ~/Library/LaunchAgents/com.edictflow.agent.plist

# Linux
systemctl --user stop edictflow-agent
systemctl --user disable edictflow-agent
rm ~/.config/systemd/user/edictflow-agent.service

# Windows
Unregister-ScheduledTask -TaskName "Edictflow Agent" -Confirm:$false
```
