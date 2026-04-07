# Vielish

English learning web app for Vietnamese speakers. Vocabulary (flashcard + SRS) and listening practice.

## Tech Stack

- **Backend:** Go (Gin or Echo) — REST API, JWT auth
- **Frontend:** Next.js (App Router)
- **Database:** PostgreSQL + Redis
- **Audio:** Cloud TTS (Google or Azure)
- **Dev:** Docker Compose

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

## Document Map
### Backend Docs
- [Project structure](./server/docs/project_structure.md)
- [API Design Standard](./server/docs/api-design-standard.md)
- [Logging Standard](./server/docs/logging-standard.md)
