# Deployment Guide

## Production Deployment

### Prerequisites

- macOS system with administrative access
- OmniFocus 4 installed and configured
- Secure token for API authentication
- Git for cloning the repository

### Quick Production Setup

```bash
# 1. Clone repository
git clone https://github.com/sho7650/omnidrop.git
cd omnidrop

# 2. Configure production environment
cp .env.example .env
# Edit .env and set secure TOKEN
vim .env

# 3. Install with service management
make install

# 4. Verify installation
make status
make logs
```

## Detailed Installation Steps

### 1. System Preparation

```bash
# Create required directories
mkdir -p ~/bin
mkdir -p ~/.local/share/omnidrop
mkdir -p ~/.local/log/omnidrop
mkdir -p ~/.config/omnidrop

# Verify OmniFocus is installed
osascript -e 'tell application "System Events" to name of every application process' | grep OmniFocus
```

### 2. Security Configuration

#### Generate Secure Token
```bash
# Generate cryptographically secure token
openssl rand -base64 32
# Example output: Kx7Hs9Lm2Np4Qr6St8Uv0Wx2Yz4Ab6Cd8Ef0Gh2Ij4=
```

#### Configure Environment
```bash
# Edit .env file
cat > .env << EOF
TOKEN=Kx7Hs9Lm2Np4Qr6St8Uv0Wx2Yz4Ab6Cd8Ef0Gh2Ij4=
EOF

# Secure the file
chmod 600 .env
```

### 3. Build and Install

```bash
# Build the binary
make build

# Install everything
make install

# This process:
# - Stops any existing service
# - Unloads the LaunchAgent
# - Installs binary to ~/bin/
# - Installs AppleScript to ~/.local/share/omnidrop/
# - Installs LaunchAgent to ~/Library/LaunchAgents/
# - Loads and starts the service
```

### 4. Service Configuration

The LaunchAgent (`com.oshiire.omnidrop.plist`) provides:

- **Automatic startup** on login
- **Restart on failure** with KeepAlive
- **Log rotation** via LaunchAgent
- **Environment management**

#### Custom LaunchAgent Configuration

```xml
<!-- Modify ~/Library/LaunchAgents/com.oshiire.omnidrop.plist -->
<dict>
    <!-- Change port if needed -->
    <key>EnvironmentVariables</key>
    <dict>
        <key>PORT</key>
        <string>8787</string>
        <key>OMNIDROP_ENV</key>
        <string>production</string>
    </dict>

    <!-- Adjust working directory -->
    <key>WorkingDirectory</key>
    <string>/Users/YOUR_USERNAME</string>
</dict>
```

### 5. Verification

```bash
# Check service status
make status
# Should show: com.oshiire.omnidrop (running)

# Check logs
make logs
# Should show: Starting OmniDrop server on port 8787

# Test the API
curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test task from deployment"}'

# Check health endpoint
curl http://localhost:8787/health
```

## Network Configuration

### Local-Only Access (Default)

The server binds to `localhost:8787` by default, accessible only from the local machine.

### Remote Access Setup

#### Option 1: SSH Tunnel (Recommended)

```bash
# From remote machine
ssh -L 8787:localhost:8787 user@mac-server.local

# Now access via localhost:8787 on remote machine
```

#### Option 2: Reverse Proxy with HTTPS

Install nginx or Caddy for HTTPS termination:

**Nginx Configuration:**
```nginx
server {
    listen 443 ssl;
    server_name omnidrop.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8787;
        proxy_set_header Authorization $http_authorization;
        proxy_set_header Content-Type $content_type;
    }
}
```

**Caddy Configuration:**
```caddy
omnidrop.yourdomain.com {
    reverse_proxy localhost:8787
}
```

### Firewall Configuration

If exposing to network:

```bash
# macOS firewall (if needed for remote access)
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /Users/YOUR_USERNAME/bin/omnidrop-server
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --unblockapp /Users/YOUR_USERNAME/bin/omnidrop-server
```

## Monitoring and Maintenance

### Log Management

```bash
# View recent logs
make logs

# Follow logs in real-time
make logs-follow

# Log locations
~/.local/log/omnidrop/stdout.log  # Standard output
~/.local/log/omnidrop/stderr.log  # Error output

# Set up log rotation (optional)
cat > ~/Library/LaunchAgents/com.omnidrop.logrotate.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.omnidrop.logrotate</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/bin/find</string>
        <string>~/.local/log/omnidrop</string>
        <string>-name</string>
        <string>*.log</string>
        <string>-mtime</string>
        <string>+7</string>
        <string>-delete</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>3</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
</dict>
</plist>
EOF

launchctl load ~/Library/LaunchAgents/com.omnidrop.logrotate.plist
```

### Health Monitoring

```bash
# Simple health check script
cat > check_omnidrop.sh << 'EOF'
#!/bin/bash
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8787/health)
if [ "$response" != "200" ]; then
    echo "OmniDrop health check failed: HTTP $response"
    # Restart service
    launchctl stop com.oshiire.omnidrop
    launchctl start com.oshiire.omnidrop
fi
EOF

chmod +x check_omnidrop.sh

# Add to crontab
crontab -e
# Add: */5 * * * * /path/to/check_omnidrop.sh
```

### Performance Monitoring

```bash
# Monitor resource usage
ps aux | grep omnidrop-server

# Check port connections
lsof -i :8787

# Monitor API response time
time curl -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Performance test"}'
```

## Backup and Recovery

### Configuration Backup

```bash
# Backup script
cat > backup_omnidrop.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="$HOME/omnidrop-backups/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup configuration
cp ~/.env "$BACKUP_DIR/"
cp ~/.local/share/omnidrop/omnidrop.applescript "$BACKUP_DIR/"
cp ~/Library/LaunchAgents/com.oshiire.omnidrop.plist "$BACKUP_DIR/"

echo "Backup completed to $BACKUP_DIR"
EOF

chmod +x backup_omnidrop.sh
./backup_omnidrop.sh
```

### Recovery Procedure

```bash
# Stop service
make stop

# Restore from backup
BACKUP_DIR="$HOME/omnidrop-backups/20240101"
cp "$BACKUP_DIR/.env" ~/
cp "$BACKUP_DIR/omnidrop.applescript" ~/.local/share/omnidrop/
cp "$BACKUP_DIR/com.oshiire.omnidrop.plist" ~/Library/LaunchAgents/

# Restart service
make start
```

## Troubleshooting Production Issues

### Service Won't Start

```bash
# Check LaunchAgent status
launchctl list | grep omnidrop

# Check for port conflicts
lsof -i :8787

# Verify file permissions
ls -la ~/bin/omnidrop-server
ls -la ~/.local/share/omnidrop/omnidrop.applescript

# Check system logs
log show --predicate 'process == "omnidrop-server"' --last 1h
```

### High Memory/CPU Usage

```bash
# Monitor resource usage
top -pid $(pgrep omnidrop-server)

# Check for runaway AppleScript processes
ps aux | grep osascript

# Restart service
make stop && make start
```

### API Not Responding

```bash
# Test local connectivity
telnet localhost 8787

# Check firewall
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --listapps

# Verify token
echo $TOKEN

# Test with verbose curl
curl -v -X POST http://localhost:8787/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Debug test"}'
```

## Security Best Practices

### Token Management

1. **Rotate tokens regularly** (monthly recommended)
   ```bash
   # Generate new token
   NEW_TOKEN=$(openssl rand -base64 32)

   # Update .env
   sed -i '' "s/TOKEN=.*/TOKEN=$NEW_TOKEN/" .env

   # Restart service
   make stop && make start
   ```

2. **Use environment-specific tokens**
   - Different tokens for dev/staging/production
   - Never commit tokens to version control

3. **Secure storage**
   ```bash
   # Ensure proper permissions
   chmod 600 .env
   chmod 700 ~/.config/omnidrop
   ```

### Network Security

1. **Use HTTPS in production** via reverse proxy
2. **Implement rate limiting** at proxy level
3. **IP whitelisting** if limited client set
4. **Monitor access logs** for anomalies

### System Security

1. **Run as non-privileged user** (default)
2. **Keep macOS updated** for latest security patches
3. **Enable FileVault** for disk encryption
4. **Regular security audits**

## Scaling Considerations

### Current Limitations

- Single instance per machine
- Sequential task processing
- Local AppleScript execution

### Future Enhancements

1. **Request Queue**: Implement task queue for batch processing
2. **Multiple Instances**: Run on multiple ports for load distribution
3. **Metrics Collection**: Prometheus/Grafana integration
4. **API Gateway**: Kong/Traefik for advanced routing

## Maintenance Schedule

### Daily
- Monitor logs for errors
- Check health endpoint

### Weekly
- Review API usage patterns
- Clean old log files
- Verify backup execution

### Monthly
- Rotate authentication tokens
- Update dependencies
- Review security logs

### Quarterly
- Full system backup
- Performance analysis
- Security audit

## Uninstallation

If needed to completely remove OmniDrop:

```bash
# Use Makefile uninstall
make uninstall

# This removes:
# - LaunchAgent service
# - Binary from ~/bin/
# - AppleScript from ~/.local/share/omnidrop/
# - Logs from ~/.local/log/omnidrop/
# - Configuration from ~/.config/omnidrop/

# Manual cleanup (if needed)
rm -rf ~/omnidrop
rm ~/.env
```