# Lamsza Platform

A high-performance, culturally authentic startpage and directory for the Szekely region.

## Overview
Lamsza is a full-stack platform: a local directory of services and websites, bilingual search, weather, news, events, and a daily quote ("mondás"). Signed-in people have a Fiók page, settings, and favorite places. The published product version is the first entry in `src/lib/publicChangelog.js` (currently v1.2.0). The footer and `/valtozasnaplo` both read that list.

## Tech Stack
- **Frontend**: SvelteKit (Vanilla CSS & JS)
- **Backend**: Go (net/http, PostgreSQL)
- **Database**: PostgreSQL (with Full-Text Search)
- **Deployment**: Docker Compose

## Quick Start
1. Configure your `.env` file (see `.env.example` if available).
2. Use the standardized service management workflow:
   ```bash
   npm run restart
   ```
   *This will start the Database, Backend, and Frontend in separate manageable processes.*

## Development
- **Start All**: `./scripts/restart_all.sh`
- **Stop All**: `pkill -f vite && pkill -f "go run main.go" && docker compose down`
- **View Logs**:
  - `tail -f backend/server_backend.log`
  - `tail -f server_frontend.log`

## Architecture
See [architecture_overview.md](architecture_overview.md) for a detailed breakdown.
