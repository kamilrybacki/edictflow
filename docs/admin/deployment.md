# Server Deployment

Deploy Claudeception server using Docker Compose, Kubernetes, or manual installation.

## Deployment Modes

Claudeception supports two deployment modes:

| Mode | Description | Use Case |
|------|-------------|----------|
| **Master-Worker** | Separate API and WebSocket processes | Production, horizontal scaling |
| **Combined (Legacy)** | Single process with both API and WebSocket | Development, small deployments |

## Docker Compose (Recommended)

The easiest way to deploy Claudeception for small to medium teams.

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 2 GB RAM minimum

### Master-Worker Mode (Production)

```bash
# Clone repository
git clone https://github.com/kamilrybacki/claudeception.git
cd claudeception
```

Create `.env` file:

```bash
# Required
DB_PASSWORD=your-secure-database-password
JWT_SECRET=your-256-bit-secret-key

# Optional
DB_USER=claudeception
DB_NAME=claudeception
REDIS_URL=redis://redis:6379/0
```

The default `docker-compose.yml` deploys master-worker architecture:

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER:-claudeception}
      POSTGRES_PASSWORD: ${DB_PASSWORD:?Set DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME:-claudeception}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U claudeception"]
      interval: 5s
      timeout: 5s
      retries: 5

  master:
    build:
      context: .
      dockerfile: server/Dockerfile
      target: master
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://${DB_USER:-claudeception}:${DB_PASSWORD}@db:5432/${DB_NAME:-claudeception}?sslmode=disable
      REDIS_URL: redis://redis:6379/0
      JWT_SECRET: ${JWT_SECRET:?Set JWT_SECRET}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  worker:
    build:
      context: .
      dockerfile: server/Dockerfile
      target: worker
    ports:
      - "8081:8081"
    environment:
      REDIS_URL: redis://redis:6379/0
      JWT_SECRET: ${JWT_SECRET}
      WORKER_PORT: "8081"
    depends_on:
      redis:
        condition: service_healthy
    deploy:
      replicas: 2
    restart: unless-stopped

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://master:8080
      NEXT_PUBLIC_WS_URL: ws://worker:8081
    depends_on:
      - master
      - worker
    restart: unless-stopped

volumes:
  postgres_data:
```

Deploy:

```bash
docker compose up -d
```

### Scaling Workers

Scale workers horizontally:

```bash
docker compose up -d --scale worker=4
```

### With TLS (Production)

Add a reverse proxy like Traefik or nginx for TLS termination:

```yaml
services:
  traefik:
    image: traefik:v2.10
    command:
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt

  master:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.api.entrypoints=websecure"
      - "traefik.http.routers.api.tls.certresolver=letsencrypt"

  worker:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.ws.rule=Host(`ws.yourdomain.com`)"
      - "traefik.http.routers.ws.entrypoints=websecure"
      - "traefik.http.routers.ws.tls.certresolver=letsencrypt"

  web:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.web.rule=Host(`app.yourdomain.com`)"
      - "traefik.http.routers.web.entrypoints=websecure"
      - "traefik.http.routers.web.tls.certresolver=letsencrypt"

volumes:
  letsencrypt:
```

## Kubernetes

For production deployments requiring high availability and scale.

### Prerequisites

- Kubernetes 1.25+
- kubectl configured
- Helm 3.0+ (optional)

### Manifests

Create namespace:

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: claudeception
```

ConfigMap and Secrets:

```yaml
# config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: claudeception-secrets
  namespace: claudeception
type: Opaque
stringData:
  db-password: your-secure-password
  jwt-secret: your-256-bit-secret
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: claudeception-config
  namespace: claudeception
data:
  DB_USER: claudeception
  DB_NAME: claudeception
  SERVER_PORT: "8080"
  WORKER_PORT: "8081"
  REDIS_URL: redis://redis:6379/0
```

Redis Deployment:

```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: claudeception
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: claudeception
spec:
  selector:
    app: redis
  ports:
    - port: 6379
```

Master Deployment:

```yaml
# master.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: master
  namespace: claudeception
spec:
  replicas: 2
  selector:
    matchLabels:
      app: master
  template:
    metadata:
      labels:
        app: master
    spec:
      containers:
        - name: master
          image: ghcr.io/kamilrybacki/claudeception-master:latest
          ports:
            - containerPort: 8080
          env:
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: claudeception-secrets
                  key: db-password
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: claudeception-secrets
                  key: jwt-secret
            - name: DATABASE_URL
              value: postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable
            - name: REDIS_URL
              valueFrom:
                configMapKeyRef:
                  name: claudeception-config
                  key: REDIS_URL
          envFrom:
            - configMapRef:
                name: claudeception-config
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: master
  namespace: claudeception
spec:
  selector:
    app: master
  ports:
    - port: 8080
```

Worker Deployment:

```yaml
# worker.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: worker
  namespace: claudeception
spec:
  replicas: 3
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
        - name: worker
          image: ghcr.io/kamilrybacki/claudeception-worker:latest
          ports:
            - containerPort: 8081
          env:
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: claudeception-secrets
                  key: jwt-secret
            - name: REDIS_URL
              valueFrom:
                configMapKeyRef:
                  name: claudeception-config
                  key: REDIS_URL
            - name: WORKER_PORT
              valueFrom:
                configMapKeyRef:
                  name: claudeception-config
                  key: WORKER_PORT
          readinessProbe:
            httpGet:
              path: /health
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: worker
  namespace: claudeception
spec:
  selector:
    app: worker
  ports:
    - port: 8081
  sessionAffinity: ClientIP  # Sticky sessions for WebSocket
```

Ingress:

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: claudeception
  namespace: claudeception
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
spec:
  tls:
    - hosts:
        - api.yourdomain.com
        - ws.yourdomain.com
        - app.yourdomain.com
      secretName: claudeception-tls
  rules:
    - host: api.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: master
                port:
                  number: 8080
    - host: ws.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: worker
                port:
                  number: 8081
    - host: app.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: web
                port:
                  number: 3000
```

Apply manifests:

```bash
kubectl apply -f namespace.yaml
kubectl apply -f config.yaml
kubectl apply -f redis.yaml
kubectl apply -f postgres.yaml
kubectl apply -f master.yaml
kubectl apply -f worker.yaml
kubectl apply -f ingress.yaml
```

## Manual Installation

For custom environments or when containers aren't available.

### Prerequisites

- Go 1.24+
- PostgreSQL 14+
- Redis 7+
- Node.js 20+ (for Web UI)

### Build Binaries

```bash
cd server

# Build master
go build -o claudeception-master ./cmd/master

# Build worker
go build -o claudeception-worker ./cmd/worker
```

### Configure

Create `/etc/claudeception/master.env`:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/claudeception?sslmode=require
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=your-256-bit-secret
SERVER_PORT=8080
```

Create `/etc/claudeception/worker.env`:

```bash
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=your-256-bit-secret
WORKER_PORT=8081
```

### Run as Services

Create systemd unit `/etc/systemd/system/claudeception-master.service`:

```ini
[Unit]
Description=Claudeception Master (API)
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=claudeception
Group=claudeception
WorkingDirectory=/opt/claudeception
EnvironmentFile=/etc/claudeception/master.env
ExecStart=/opt/claudeception/claudeception-master
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Create systemd unit `/etc/systemd/system/claudeception-worker.service`:

```ini
[Unit]
Description=Claudeception Worker (WebSocket)
After=network.target redis.service

[Service]
Type=simple
User=claudeception
Group=claudeception
WorkingDirectory=/opt/claudeception
EnvironmentFile=/etc/claudeception/worker.env
ExecStart=/opt/claudeception/claudeception-worker
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable claudeception-master claudeception-worker
sudo systemctl start claudeception-master claudeception-worker
```

## Post-Deployment

After deployment:

1. **Run migrations**: Migrations run automatically on master start
2. **Create admin user**: Use the API or seed script
3. **Configure OAuth**: Set up SSO providers
4. **Test connectivity**: Verify agents can connect to workers

## Health Checks

### Master Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Basic health check |
| `GET /` | API status |

### Worker Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Worker health with agent stats |

Example worker health response:

```json
{
  "status": "ok",
  "agents": 5,
  "teams": 2,
  "subscriptions": 2
}
```

## Troubleshooting

### Redis Connection Failed

```bash
# Check Redis is running
docker compose ps redis

# Test connection
redis-cli -u $REDIS_URL ping
```

### Workers Not Receiving Events

```bash
# Check Redis pub/sub
redis-cli subscribe "team:*:rules"

# In another terminal, publish test event
redis-cli publish "team:test:rules" '{"event":"test"}'
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker compose ps db

# Check connection string
docker compose exec master env | grep DATABASE_URL

# Test connection manually
psql $DATABASE_URL -c "SELECT 1"
```

### Migrations Failed

```bash
# Check migration status
migrate -path migrations -database "$DATABASE_URL" version

# Force to specific version if needed
migrate -path migrations -database "$DATABASE_URL" force VERSION
```

### WebSocket Connections Failing

Ensure your reverse proxy supports WebSocket upgrades:

```nginx
# nginx configuration for workers
location /ws {
    proxy_pass http://worker:8081;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;
}
```

### Agent Can't Connect

1. Verify worker is healthy: `curl http://worker:8081/health`
2. Check JWT secret matches between master and worker
3. Verify network connectivity from agent to worker
4. Check firewall rules for WebSocket port
