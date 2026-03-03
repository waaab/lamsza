# Architecture & Deployment Guide

## Core Architecture
- **Frontend**: SvelteKit, compiled to a Node.js adapter build.
- **Backend**: Go (1.25.7) compiling to a single binary `sz-gugel-bin`.
- **Database**: PostgreSQL.

## Environment Variables
The application relies on a `.env` file at the root of the project. This file is **not** committed to version control.

### Required Keys:
- `PORT`: The port the Go backend listens on (default: `3000`).
- `DATABASE_URL`: The PostgreSQL connection string. (e.g. `postgres://lamsza_user:lamsza_password@localhost:5433/lamsza?sslmode=disable`)
- `VITE_API_BASE_URL`: The base URL for the backend API used by SvelteKit (e.g. `http://localhost:3000`).
- `VITE_WEATHER_API_KEY`: OpenWeatherMap API key for weather data.

## Continuous Integration (CI)
We use GitHub Actions for continuous integration.
- **Workflow**: `.github/workflows/build-test.yml`
- **Trigger**: Runs on every `push`.
- **Function**: Verifies that both the SvelteKit frontend and Go backend compile successfully without errors.
- **Note**: This workflow does **not** deploy code to the DigitalOcean server. It is strictly for build and test verification.

## Local Development
1. Ensure your `.env` is configured.
2. Start the backend: `cd backend && go run main.go`
3. Start the frontend: `npm run dev`
