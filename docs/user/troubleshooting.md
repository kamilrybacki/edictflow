# Troubleshooting

Common issues and solutions for the Edictflow agent.

## Connection Issues

### Agent Won't Connect

**Symptoms:**
- `edictflow-agent status` shows `Connected: No`
- Timeout errors during login or sync

**Solutions:**

1. **Check server URL:**
   ```bash
   edictflow-agent config show | grep server
   ```
   Verify the URL is correct.

2. **Test network connectivity:**
   ```bash
   curl -I https://your-server.example.com/health
   ```
   Should return HTTP 200.

3. **Check firewall:**
   Ensure outbound HTTPS (443) and WSS are allowed.

4. **Try re-authentication:**
   ```bash
   edictflow-agent logout
   edictflow-agent login https://your-server.example.com
   ```

5. **Check proxy settings:**
   ```bash
   export HTTPS_PROXY=http://proxy.example.com:8080
   edictflow-agent start
   ```

### WebSocket Connection Drops

**Symptoms:**
- Frequent reconnections
- "Connection lost" notifications

**Solutions:**

1. **Check network stability:**
   Unstable WiFi or VPN can cause drops.

2. **Increase timeout:**
   ```bash
   edictflow-agent config set websocket.timeout 60s
   ```

3. **Check server logs:**
   Ask admin to check for server-side issues.

---

## Authentication Issues

### Login Fails

**Symptoms:**
- "Authentication failed" error
- Device code not accepted

**Solutions:**

1. **Check server URL:**
   Ensure you're using the correct server address.

2. **Try different browser:**
   Some corporate browsers block OAuth flows.

3. **Clear browser cookies:**
   For the Edictflow server domain.

4. **Check SSO configuration:**
   Contact admin if OAuth provider is misconfigured.

### Token Expired

**Symptoms:**
- "Unauthorized" errors
- Agent stops syncing

**Solutions:**

```bash
edictflow-agent logout
edictflow-agent login https://your-server.example.com
edictflow-agent restart
```

---

## File Watching Issues

### Changes Not Detected

**Symptoms:**
- Editing files doesn't trigger enforcement
- No change events logged

**Solutions:**

1. **Check file is covered by rule:**
   ```bash
   edictflow-agent rules
   ```
   Verify triggers match your file path.

2. **Use polling mode:**
   For network filesystems or containers:
   ```bash
   edictflow-agent start --poll-interval 500ms
   ```

3. **Check inotify limits (Linux):**
   ```bash
   cat /proc/sys/fs/inotify/max_user_watches
   # Increase if needed
   sudo sysctl fs.inotify.max_user_watches=65536
   ```

4. **Restart agent:**
   ```bash
   edictflow-agent restart
   ```

### Excessive CPU Usage

**Symptoms:**
- Agent using high CPU
- System slowdown

**Solutions:**

1. **Check for recursive watches:**
   Watching large directories can cause high CPU.

2. **Use polling with longer interval:**
   ```bash
   edictflow-agent start --poll-interval 2s
   ```

3. **Exclude unnecessary paths:**
   Contact admin to refine rule triggers.

---

## Rule Issues

### Rules Not Syncing

**Symptoms:**
- `edictflow-agent rules` shows no rules
- Rules out of date

**Solutions:**

1. **Force sync:**
   ```bash
   edictflow-agent sync --force
   ```

2. **Check team assignment:**
   ```bash
   edictflow-agent status
   ```
   Verify you're assigned to the correct team.

3. **Clear cache:**
   ```bash
   # macOS
   rm -rf ~/Library/Application\ Support/edictflow/cache/

   # Linux
   rm -rf ~/.config/edictflow/cache/

   edictflow-agent restart
   ```

### Wrong Rule Applied

**Symptoms:**
- Unexpected enforcement on files
- Wrong content being enforced

**Solutions:**

1. **Check rule triggers:**
   ```bash
   edictflow-agent rules --json | jq '.[] | {name, triggers}'
   ```

2. **Check rule priority:**
   Multiple rules may match; contact admin for priority settings.

3. **Verify file path:**
   Ensure the path matches expected triggers.

---

## Daemon Issues

### Agent Won't Start

**Symptoms:**
- `edictflow-agent start` fails
- No process running

**Solutions:**

1. **Check if already running:**
   ```bash
   edictflow-agent status
   # If stuck, force stop:
   edictflow-agent stop --force
   ```

2. **Check PID file:**
   ```bash
   # macOS
   rm ~/Library/Application\ Support/edictflow/agent.pid

   # Linux
   rm ~/.config/edictflow/agent.pid
   ```

3. **Check logs:**
   ```bash
   edictflow-agent start --foreground
   # View output for errors
   ```

4. **Check permissions:**
   ```bash
   ls -la ~/.config/edictflow/
   # Should be owned by your user
   ```

### Agent Crashes

**Symptoms:**
- Agent stops unexpectedly
- Need to restart frequently

**Solutions:**

1. **Check logs for errors:**
   ```bash
   # macOS
   cat ~/Library/Logs/edictflow-agent.log

   # Linux
   cat ~/.local/share/edictflow/agent.log
   ```

2. **Run in foreground to debug:**
   ```bash
   edictflow-agent start --foreground
   ```

3. **Update to latest version:**
   ```bash
   edictflow-agent update
   ```

4. **Report issue:**
   If crashes persist, file a GitHub issue with logs.

---

## Data Issues

### Corrupted Database

**Symptoms:**
- SQLite errors in logs
- Agent won't start with database errors

**Solutions:**

1. **Reset local database:**
   ```bash
   # macOS
   rm ~/Library/Application\ Support/edictflow/agent.db

   # Linux
   rm ~/.config/edictflow/agent.db

   edictflow-agent restart
   # Will re-sync from server
   ```

### Missing Configuration

**Symptoms:**
- Config values not persisting
- Default behavior despite configuration

**Solutions:**

1. **Check config location:**
   ```bash
   edictflow-agent config show
   ```

2. **Reset config:**
   ```bash
   edictflow-agent config reset
   ```

3. **Re-apply settings:**
   ```bash
   edictflow-agent config set <key> <value>
   ```

---

## Platform-Specific Issues

### macOS

#### Gatekeeper Blocks Agent

```bash
xattr -d com.apple.quarantine /usr/local/bin/edictflow-agent
```

#### Accessibility Permissions

If notifications don't work:

1. System Preferences → Security & Privacy → Privacy
2. Enable "Notifications" for Edictflow

### Linux

#### SELinux Blocks Agent

```bash
# Check if SELinux is blocking
ausearch -m avc -ts recent

# Allow if needed
setsebool -P allow_execstack 1
```

#### AppArmor Issues

```bash
# Disable AppArmor profile for agent
sudo ln -s /etc/apparmor.d/edictflow-agent /etc/apparmor.d/disable/
sudo apparmor_parser -R /etc/apparmor.d/edictflow-agent
```

### Windows

#### Firewall Blocking

1. Windows Defender Firewall → Allow an app
2. Add `edictflow-agent.exe`
3. Allow both private and public networks

#### Antivirus False Positive

Add exception for:
- `%LOCALAPPDATA%\Edictflow\`
- `edictflow-agent.exe`

---

## Getting Help

### Collect Diagnostics

```bash
edictflow-agent status --json > status.json
edictflow-agent rules --json > rules.json
edictflow-agent changes --limit 50 --json > changes.json

# macOS
cp ~/Library/Logs/edictflow-agent.log ./agent.log

# Linux
cp ~/.local/share/edictflow/agent.log ./agent.log

# Create archive
tar czf edictflow-diagnostics.tar.gz status.json rules.json changes.json agent.log
```

### Contact Support

1. **GitHub Issues**: [github.com/kamilrybacki/edictflow/issues](https://github.com/kamilrybacki/edictflow/issues)
2. **Include diagnostics** with your issue
3. **Describe steps to reproduce**

### Useful Log Levels

Enable debug logging for detailed information:

```bash
edictflow-agent config set log.level debug
edictflow-agent restart
```

Remember to reset after troubleshooting:

```bash
edictflow-agent config set log.level info
```
