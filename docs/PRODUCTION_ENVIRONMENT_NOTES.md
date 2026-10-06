# LAMSZA NETWORK — COMPLETE INFRASTRUCTURE, CONFIGURATION & DEPLOYMENT RUNBOOK

## 1. INFRASTRUCTURE & NETWORK OVERVIEW

- **Platform**: DigitalOcean
- **Project Name**: `Lamsza-Network`
- **Region / Datacenter**: FRA1 (Frankfurt)
- **Droplet Specs**: 2 GB RAM / 1 vCPU / 50 GB NVMe Disk / Ubuntu 24.04 LTS
- **Dedicated Reserved IPv4**: `146.190.204.232`
- **Primary Deploy User**: `attila` (sudo group, full ownership of `/var/www/`)
- **Swap Configuration**: 2 GB swap active at `/swapfile` (`vm.swappiness=10`)

### DNS Configuration (DigitalOcean DNS)

- `lamsza.com` → Active on **Google Cloud AppEngine** (pending migration to this droplet).
- `szotar.lamsza.com` → Points to `146.190.204.232`.
- `jatszoter.lamsza.com` → Points to `146.190.204.232`.
- `admin.lamsza.com` → Points to `146.190.204.232`.
  - **Phase 2:** DNS/Nginx/TLS/`admin.service` placeholder exist. The independent `lamsza-admin` app runs locally first (`http://localhost:5173` / API `:3000`). Production cutover (static to `/var/www/admin/public`, Go on `:8083`) is deferred. Szótár/Játszótér `/admin` stay in their own apps for now.
- **Decommissioned & Purged**: Legacy droplet `46.101.116.17` deleted; unused DNS records (`api.lamsza.com`, `accounts.lamsza.com`, and all `*-staging` domains) purged.

## 2. SECURITY & OS HARDENING

### SSH Hardening (`/etc/ssh/sshd_config.d/`)

- Port: `22`
- Root login: Disabled (`PermitRootLogin no`)
- Password authentication: Disabled (`PasswordAuthentication no`)
- Public key authentication: Enforced for `attila`

### UFW (Uncomplicated Firewall)

- Default Incoming: `Deny`
- Default Outgoing: `Allow`
- Allowed Ports:
  - `22/tcp` (SSH) — IPv4 & IPv6
  - `80/tcp` (HTTP) — IPv4 & IPv6
  - `443/tcp` (HTTPS) — IPv4 & IPv6

### Automated Patching & System Auditing

- **Fail2ban**: Active with standard jail monitoring SSH authentication failures.
- **Unattended Upgrades**: Enabled for automatic security patch installations.
- **Systemd Journal Storage**: Capped at `100M` in `/etc/systemd/journald.conf` (`SystemMaxUse=100M`).

## 3. DATABASE CONFIGURATION (POSTGRESQL)

PostgreSQL is installed and configured strictly for local loopback connections. External network listening is disabled.

- **Listen Addresses**: `127.0.0.1:5432`, `[::1]:5432`
- **Databases & User Isolation**:
  - **Main App**: Database `lamsza` | User `lamsza_user`
  - **Szótár**: Database `szotar` | User `szotar_user`
  - **Játszótér**: Database `jatszoter` | User `jatszoter_user`
- **Internal Connection String Syntax**:

  ```
  postgres://<user>:<password>@127.0.0.1:5432/<dbname>?sslmode=disable
  ```

## 4. WEB SERVER & REVERSE PROXY (NGINX + LET'S ENCRYPT)

### Global Security Headers (`/etc/nginx/snippets/security-headers.conf`)

Included across all server blocks:

```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "geolocation=(), camera=(), microphone=()" always;
```

### Content-Security-Policy

The Lámsza portal ships its own policy inside every built page as a
`<meta http-equiv="content-security-policy">` tag (`kit.csp` in `svelte.config.js`), so the site is
covered without any Nginx change. Keep `X-Frame-Options` above: `frame-ancestors` has no effect
from a `meta` tag.

Two prerendered redirect stubs, `/profil` and `/szek`, carry no policy because SvelteKit writes
them as a bare redirect. They hold no user content, and their `<meta http-equiv="refresh">`
fallback still forwards the visitor if the inline redirect is blocked.

**Still to do (not yet applied on the droplet):** add the same policy as a response header so the
stubs and any future non-SvelteKit page are covered too, and so the policy survives a page that
does not go through SvelteKit. Take the exact value from a built page:

```bash
grep -o 'content-security-policy" content="[^"]*"' /var/www/lamsza/public/index.html
```

The admin, Szótár and Játszótér sites have no policy of their own yet; until they do, a header in
their server blocks is the only cover they get.

### Server Token Masking

`server_tokens off;` enabled inside `/etc/nginx/nginx.conf`.

### Routing model (verified)

Strict split for `szotar`, `jatszoter`, and `admin` (Nginx syntax OK, reloaded, live-tested):

- **`/api/`** → local Go upstream (`8081` / `8082` / `8083`)
- **`/`** → static files under `public/`, SPA fallback via `try_files` → **`/app.html`**

Matches SvelteKit `adapter-static` `fallback: 'app.html'`. Deploy `frontend/dist/*` (or lamsza `dist/*`) straight into `/var/www/<app>/public/` — **no rename to `index.html`**.

**Live SPA checks (HTTP 200, served from `app.html`):**

- `https://szotar.lamsza.com/szo/kutya`
- `https://jatszoter.lamsza.com/jatszok/szokereso`
- `https://admin.lamsza.com/dashboard/settings`

**API boundary:** `/api/` is proxied to Go (not caught by SPA). A `502` when the Go process is down (or the health path does not exist) still confirms the proxy path; app health endpoints are typically `/api/health` (not `/api/v1/health`).

### Active Nginx Site Configurations (`/etc/nginx/sites-available/` → `/etc/nginx/sites-enabled/`)

#### A. Szótár (`/etc/nginx/sites-available/szotar.lamsza.com`)

```nginx
server {
    server_name szotar.lamsza.com;
    root /var/www/szotar/public;
    index app.html;

    include snippets/security-headers.conf;

    location /api/ {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /app.html;
    }

    listen 443 ssl;
    listen [::]:443 ssl;
    ssl_certificate /etc/letsencrypt/live/szotar.lamsza.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/szotar.lamsza.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}

server {
    listen 80;
    listen [::]:80;
    server_name szotar.lamsza.com;
    return 301 https://$host$request_uri;
}
```

#### B. Játszótér (`/etc/nginx/sites-available/jatszoter.lamsza.com`)

```nginx
server {
    server_name jatszoter.lamsza.com;
    root /var/www/jatszoter/public;
    index app.html;

    include snippets/security-headers.conf;

    location /api/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /app.html;
    }

    listen 443 ssl;
    listen [::]:443 ssl;
    # Shared dual-SAN cert: Domains = szotar.lamsza.com jatszoter.lamsza.com
    ssl_certificate /etc/letsencrypt/live/szotar.lamsza.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/szotar.lamsza.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}

server {
    listen 80;
    listen [::]:80;
    server_name jatszoter.lamsza.com;
    return 301 https://$host$request_uri;
}
```

#### C. Admin (`/etc/nginx/sites-available/admin.lamsza.com`)

```nginx
server {
    server_name admin.lamsza.com;
    root /var/www/admin/public;
    index app.html;

    include snippets/security-headers.conf;

    location /api/ {
        proxy_pass http://127.0.0.1:8083;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /app.html;
    }

    listen 443 ssl;
    listen [::]:443 ssl;
    ssl_certificate /etc/letsencrypt/live/admin.lamsza.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.lamsza.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}

server {
    listen 80;
    listen [::]:80;
    server_name admin.lamsza.com;
    return 301 https://$host$request_uri;
}
```

#### D. Main Domain (`/etc/nginx/sites-available/lamsza.com`) — Staged

HTTP-only until DNS cutover from App Engine. Apply the same `/api/` + SPA split before go-live. Lamsza also prerenders many `.html` pages (`megyek.html`, `admin.html`, …), so prefer `$uri.html` in `try_files`:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name lamsza.com www.lamsza.com;
    root /var/www/lamsza/public;
    index app.html;

    include snippets/security-headers.conf;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 8m;
    }

    location / {
        try_files $uri $uri.html $uri/ /app.html;
    }
}
```

### SSL Certificates

Verified via `sudo certbot certificates`:

- `szotar.lamsza.com` — ECDSA; **Domains:** `szotar.lamsza.com`, `jatszoter.lamsza.com`  
  Paths: `/etc/letsencrypt/live/szotar.lamsza.com/{fullchain,privkey}.pem`
- `admin.lamsza.com` — ECDSA; **Domains:** `admin.lamsza.com`  
  Paths: `/etc/letsencrypt/live/admin.lamsza.com/{fullchain,privkey}.pem`
- **Automation**: `certbot.timer` (renewal checks twice daily)

## 5. APPLICATION RUNTIMES & SYSTEMD SERVICES

All applications execute as user `attila`, read isolated `.env` configuration files, and run with security sandboxing enabled (`ProtectSystem=full`, `ProtectHome=true`, `NoNewPrivileges=true`).

### Port and Path Mapping


| Application     | Domain                 | Port   | Working Directory    | Static Assets               | Service Unit              |
| --------------- | ---------------------- | ------ | -------------------- | --------------------------- | ------------------------- |
| **Main Portal** | `lamsza.com`           | `8080` | `/var/www/lamsza`    | `/var/www/lamsza/public`    | `lamsza.service` (staged) |
| **Szótár**      | `szotar.lamsza.com`    | `8081` | `/var/www/szotar`    | `/var/www/szotar/public`    | `szotar.service`          |
| **Játszótér**   | `jatszoter.lamsza.com` | `8082` | `/var/www/jatszoter` | `/var/www/jatszoter/public` | `jatszoter.service`       |
| **Admin**       | `admin.lamsza.com`     | `8083` | `/var/www/admin`     | `/var/www/admin/public`     | `admin.service`           |


### Systemd Service Unit Files

#### `/etc/systemd/system/szotar.service`

```ini
[Unit]
Description=Lamsza Szotar Service
After=network.target postgresql.service

[Service]
Type=simple
User=attila
Group=attila
WorkingDirectory=/var/www/szotar
ExecStart=/var/www/szotar/szotar
Restart=always
RestartSec=5s
EnvironmentFile=-/var/www/szotar/.env
Environment=PORT=8081

# Security Sandbox
ProtectSystem=full
ProtectHome=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target

```

#### `/etc/systemd/system/jatszoter.service`

```ini
[Unit]
Description=Lamsza Jatszoter Service
After=network.target postgresql.service

[Service]
Type=simple
User=attila
Group=attila
WorkingDirectory=/var/www/jatszoter
ExecStart=/var/www/jatszoter/jatszoter
Restart=always
RestartSec=5s
EnvironmentFile=-/var/www/jatszoter/.env
Environment=PORT=8082

# Security Sandbox
ProtectSystem=full
ProtectHome=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target

```

#### `/etc/systemd/system/admin.service`

```ini
[Unit]
Description=Lamsza Admin Service
After=network.target postgresql.service

[Service]
Type=simple
User=attila
Group=attila
WorkingDirectory=/var/www/admin
ExecStart=/var/www/admin/admin
Restart=always
RestartSec=5s
EnvironmentFile=-/var/www/admin/.env
Environment=PORT=8083

# Security Sandbox
ProtectSystem=full
ProtectHome=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target

```

## 6. CURSOR BUILD & DEPLOYMENT CONTRACT

Builds are handled off-server (in development/local machine or CI). The production server runs compiled standalone artifacts only.

### Compilation Target

- Architecture: Linux x86_64 (`GOOS=linux GOARCH=amd64`)
- Command:

  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o <binary_name> .
  ```

### Directory Placement Contract

- Binary: `/var/www/<app>/<binary_name>` (must have executable permission: `chmod +x`)
- Static Frontend Files: `/var/www/<app>/public/`
- Environment Variables: `/var/www/<app>/.env` (permissions `0600`)

### Service Management Commands

```bash
sudo systemctl restart szotar
sudo systemctl restart jatszoter
sudo systemctl restart admin
```

