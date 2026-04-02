# Vielish

English learning web app for Vietnamese speakers.

## Features (MVP)

- **Vocabulary** — Flashcard with spaced repetition (SM-2), topic-based learning, quiz
- **Listening** — Audio lessons with fill-in-the-blank, multiple choice, true/false exercises

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (Gin/Echo) |
| Frontend | Next.js (App Router) |
| Database | PostgreSQL + Redis |
| Audio | Cloud TTS (Google/Azure) |
| Dev | Docker Compose |

## Getting Started

```bash
# Start database
docker-compose up -d postgres redis

# Run backend (from server/)
cd server
go run cmd/api/main.go

# Run frontend (from web/)
cd web
npm install
npm run dev
```

## Project Structure

```
server/          — Go API server
  cmd/api/       — entrypoint
  internal/      — business logic (auth, vocab, listening, srs)
  pkg/           — shared utilities
  migrations/    — SQL migrations
web/             — Next.js frontend
  app/           — pages
  components/    — UI components
  lib/           — API client, utils
docs/            — specs & documentation
```

## Roadmap

- [x] Design spec
- [ ] Project scaffolding (Go + Next.js)
- [ ] Auth (register/login)
- [ ] Vocabulary module (flashcard + SRS)
- [ ] Listening module
- [ ] Progress tracking & dashboard
- [ ] AI conversation practice
- [ ] Mobile app
