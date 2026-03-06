# Lamsza Platform

A high-performance, culturally authentic startpage and directory for the Szekely region.

## Overview
Lamsza is a full-stack platform featuring a local service directory with advanced bilingual search, real-time weather, news aggregation, and a random quote ("mondás") system.

## Tech Stack
- **Frontend**: SvelteKit (Vanilla CSS & JS)
- **Backend**: Go (Gin, PostgreSQL)
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
鼓
