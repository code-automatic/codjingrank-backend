# VPS Deployment with Nginx + HTTPS

> **v2.0 migration note:** This runbook contains legacy Repository/GitHub
> deployment steps. Keep the gateway/proxy/runtime portions; do not provision
> `repo_service`, GitHub App/OAuth or webhooks for the target product. Forum and
> AI deployment steps are pending implementation.

## Architecture

```
Internet → Nginx (HTTPS :443) → 127.0.0.1:8080 (api_gateway)
                                   → /var/www/rinogen-leetcode (SPA static files)
```

- **api_gateway** runs on port **8080** (not 8000).
- API routes are proxied at **root level** (no `/api` prefix) — e.g., `/auth`, `/health`, `/subjects`, `/run`, etc.
- **Frontend SPA** is served from `/var/www/rinogen-leetcode` with `try_files` fallback to `index.html`.
- All other backend services (auth, run, submission, worker, and future Forum/AI) stay private inside Docker and are **never** exposed publicly. `repo_service` is legacy and should not be started for v2.0.

## Prerequisites

- Ubuntu 22.04+ VPS
- Docker + Docker Compose
- Domain pointed to VPS IP (e.g., code.innogenlab.com)

## 1. Clone and set up

```bash
git clone git@github.com:warmdev17/innogen-backend-microservice.git
cd innogen-backend-microservice

# Create secrets directory
mkdir -p secrets
# Place GitHub App private key:
# cp /path/to/private-key.pem secrets/github-key.pem
chmod 600 secrets/github-key.pem
```

## 2. Environment file

```bash
cp .env.example .env
# Edit .env with production values:
#   JWT_SECRET=<random-64-char>
#   GITHUB_APP_ID=3721867
#   GITHUB_PRIVATE_KEY_PATH=/app/secrets/github-key.pem
#   GITHUB_WEBHOOK_SECRET=<from-github-app-settings>
#   GITHUB_OAUTH_CLIENT_ID=Ov23li...
#   GITHUB_OAUTH_CLIENT_SECRET=<secret>
#   GITHUB_OAUTH_REDIRECT_URL=https://code.innogenlab.com/github/oauth/callback
#   GITHUB_OAUTH_FRONTEND_REDIRECT_URL=https://code.innogenlab.com/github
```

## 3. Start services

```bash
# Production (only gateway exposed)
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# Apply schema + seed
docker compose exec -T postgres psql -U innogen -d innogen < schema.sql
docker compose exec -T postgres psql -U innogen -d innogen < seeds/dev_seed.sql
docker compose exec -T postgres psql -U innogen -d innogen < seeds/dev_problem_pack.sql
```

## 4. Deploy Frontend SPA

```bash
# Create web root and deploy your frontend build
sudo mkdir -p /var/www/rinogen-leetcode
# Copy frontend build output (index.html, assets/, etc.)
sudo cp -r /path/to/frontend/dist/* /var/www/rinogen-leetcode/
sudo chown -R www-data:www-data /var/www/rinogen-leetcode
```

## 5. Nginx config

Key points:
- **API routes are proxied at root level** (no `/api` prefix). Each API path prefix gets its own `location` block that proxies to `http://127.0.0.1:8080` (api_gateway).
- **Frontend SPA** is served from `/var/www/rinogen-leetcode`. The catch-all `location /` uses `try_files` to fall back to `index.html` for client-side routing.
- **api_gateway runs on port 8080** — all API proxying goes to this port.

Create `/etc/nginx/sites-available/rinnogen`:

```nginx
server {
    listen 80;
    server_name code.innogenlab.com www.code.innogenlab.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name code.innogenlab.com www.code.innogenlab.com;

    root /var/www/rinogen-leetcode;
    index index.html index.htm;

    ssl_certificate /etc/nginx/ssl/innogenlab.com.pem;
    ssl_certificate_key /etc/nginx/ssl/innogenlab.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    client_max_body_size 2M;

    # Backend API — proxy to api_gateway
    location /health { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /auth { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /subjects { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /sessions { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /lessons { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /problems { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /run { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /submit { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /submissions { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /me { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /github { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /admin { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /docs { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /webhooks { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }
    location /repositories { proxy_pass http://127.0.0.1:8080; proxy_http_version 1.1; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; proxy_set_header X-Forwarded-Proto $scheme; }

    # Frontend SPA — everything else
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

Enable:

```bash
sudo ln -sf /etc/nginx/sites-available/rinnogen /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 6. HTTPS with Certbot

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d code.innogenlab.com -d www.code.innogenlab.com
```

Auto-renewal:

```bash
sudo certbot renew --dry-run  # test
# Auto-renews via systemd timer
```

## 7. GitHub App Production URLs

After HTTPS is set up, update GitHub App settings:
- **Callback URL**: `https://code.innogenlab.com/github/oauth/callback`
- **Webhook URL**: `https://code.innogenlab.com/webhooks/github`
- **Setup URL**: `https://code.innogenlab.com/github`

## 8. Firewall

```bash
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP (for certbot)
sudo ufw allow 443   # HTTPS
sudo ufw deny 8080   # Block direct gateway access
sudo ufw deny 8081   # Block auth_service
sudo ufw deny 8082   # Block run_service
sudo ufw deny 8083   # Block submission_service
sudo ufw deny 8084   # Block repo_service
sudo ufw deny 5432   # Block postgres
sudo ufw deny 6379   # Block redis
sudo ufw deny 2000   # Block piston
sudo ufw enable
```

## 9. Secrets mount

In `docker-compose.prod.yml`, add to repo_service:

```yaml
  repo_service:
    volumes:
      - ./secrets:/app/secrets:ro
```

Set `GITHUB_PRIVATE_KEY_PATH=/app/secrets/github-key.pem`

## 10. Useful commands

```bash
# View logs
docker compose logs -f api_gateway

# Restart after code change
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# Database backup
docker compose exec postgres pg_dump -U innogen innogen > backup.sql

# Check health
curl -s https://code.innogenlab.com/health
```

## Production Checklist

- [ ] .env secrets set (JWT, GitHub, OAuth)
- [ ] GitHub private key in secrets/
- [ ] Frontend deployed to /var/www/rinogen-leetcode
- [ ] Nginx HTTPS working
- [ ] Firewall active
- [ ] GitHub App callback/webhook URLs updated
- [ ] Health check passes
- [ ] Submission flow tested
- [ ] GitHub commit flow tested
