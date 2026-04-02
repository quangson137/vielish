# Vielish

English learning web app for Vietnamese speakers. Vocabulary (flashcard + SRS) and listening practice.

## Tech Stack

- **Backend:** Go (Gin or Echo) — REST API, JWT auth
- **Frontend:** Next.js (App Router)
- **Database:** PostgreSQL + Redis
- **Audio:** Cloud TTS (Google or Azure)
- **Dev:** Docker Compose

## Project Structure

```
server/          — Go backend
  cmd/api/       — entrypoint
  internal/      — business logic (auth, vocab, listening, srs)
  pkg/           — shared utilities
  migrations/    — SQL migrations
web/             — Next.js frontend
  app/           — pages (App Router)
  components/    — UI components
  lib/           — API client, utils
docs/            — documentation & specs
```

## Design Spec

Full spec at `docs/agent-docs/specs/2026-04-02-vielish-mvp-design.md`.

## Key Decisions

- **SRS algorithm:** SM-2 with 3-level rating: Hard (1) / OK (3) / Easy (5)
- **API prefix:** `/api/` — REST only, no WebSocket in MVP
- **Auth:** JWT tokens + Redis session
- **Language:** All user-facing content in Vietnamese, learning content in English
- **MVP scope:** Vocabulary + Listening only. No AI features in MVP.

## Commands

```bash
# Backend (from server/)
go run cmd/api/main.go

# Frontend (from web/)
npm run dev

# Database
docker-compose up -d postgres redis
```

## Conventions

- Go: follow standard Go project layout (`internal/`, `pkg/`, `cmd/`)
- Frontend: Next.js App Router conventions, components in `components/`
- API responses: JSON, consistent error format `{ "error": "message" }`
- Database migrations: numbered SQL files in `server/migrations/`
