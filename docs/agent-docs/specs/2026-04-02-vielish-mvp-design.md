# Vielish MVP — Design Spec

English learning web app for Vietnamese speakers.

## Overview

- **Target users:** Vietnamese learners of all levels (beginner → advanced)
- **Platform:** Web app (Next.js), mobile later
- **MVP features:** Vocabulary (flashcard + SRS), Listening practice
- **Business model:** TBD — focus on product first

## Tech Stack

- **Backend:** Go (Gin or Echo) — REST API, JWT auth
- **Frontend:** Next.js (App Router) — React-based SPA/SSR
- **Database:** PostgreSQL (primary data) + Redis (cache, session, SRS queue)
- **Audio:** Cloud TTS (Google or Azure) for native speaker audio
- **Infra:** Docker Compose for local dev

## Project Structure

```
vielish/
  server/               — Go backend
    cmd/api/            — entrypoint
    internal/           — business logic
      auth/
      vocab/
      listening/
      srs/
    pkg/                — shared utilities
    migrations/         — SQL migrations
  web/                  — Next.js frontend
    app/                — pages (App Router)
    components/         — UI components
    lib/                — API client, utils
  docs/                 — documentation
  docker-compose.yml
```

## Data Model

### Users

```sql
users: id, email, password_hash, display_name, level (beginner/intermediate/advanced), created_at
```

Auth via JWT + Redis session.

### Vocabulary

```sql
topics: id, name, name_vi, description, level
words: id, word, ipa_phonetic, part_of_speech, vi_meaning, en_definition, example_sentence, example_vi_translation, audio_url, image_url, level, topic_id
user_word_progress: user_id, word_id, ease_factor, interval_days, next_review_at, review_count, last_reviewed_at
```

SRS uses SM-2 algorithm. `user_word_progress` tracks per-user review state.

### Listening

```sql
listening_lessons: id, title, title_vi, audio_url, transcript, transcript_vi, level, topic_id, duration_seconds
listening_questions: id, lesson_id, question_text, question_type (fill_blank/multiple_choice/true_false), correct_answer, options (jsonb)
user_listening_progress: user_id, lesson_id, score, completed_at
```

## API Endpoints

### Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Register (email, password, display_name) |
| POST | `/api/auth/login` | Login → JWT token |
| POST | `/api/auth/refresh` | Refresh token |

### Vocabulary

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/topics` | List topics (filter by level) |
| GET | `/api/topics/:id/words` | List words in topic |
| GET | `/api/words/:id` | Word detail |
| GET | `/api/review/due` | Words due for SRS review today |
| POST | `/api/review/:wordId` | Submit review rating: 1 (Hard), 3 (OK), 5 (Easy) |
| GET | `/api/quiz/:topicId` | Get quiz questions for topic |
| POST | `/api/quiz/:topicId` | Submit quiz answers |

### Listening

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/listening/lessons` | List lessons (filter by level, topic) |
| GET | `/api/listening/lessons/:id` | Lesson detail (audio_url, questions) |
| POST | `/api/listening/lessons/:id/submit` | Submit answers → score + corrections |

### Progress

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/progress/summary` | Overview: words learned, streak, accuracy |
| GET | `/api/progress/vocab` | Vocabulary progress detail |
| GET | `/api/progress/listening` | Listening progress detail |

## Core Features

### Vocabulary (Flashcard + SRS)

- **Flashcard:** Front = English word + IPA + audio. Back = Vietnamese meaning + example sentence + image.
- **SRS (SM-2):** After each review, user rates difficulty: Hard (1) / OK (3) / Easy (5). System calculates next review date.
- **Topic-based:** Users pick topics (Travel, Food, Business...), ~20-50 words per topic per level.
- **Modes:** Learn (new words) → Review (SRS-scheduled) → Quiz (multiple choice test).

### Listening Practice

- **Leveled lessons:** Each lesson has audio + transcript.
- **Question types:** Fill in the blank, multiple choice, true/false.
- **Playback speed:** 0.5x / 0.75x / 1x / 1.25x.
- **Transcript:** Hidden during exercise, revealed after answering (with Vietnamese translation).

## UI Screens

1. **Dashboard** — Progress overview (words learned, streak, due reviews), continue learning suggestions.
2. **Topic List** — Browse topics by level, see progress per topic.
3. **Flashcard View** — Card flip interaction. Front: word + IPA + audio. Back: meaning + example + difficulty rating buttons (Hard/OK/Easy).
4. **Quiz View** — Multiple choice questions per topic.
5. **Listening Player** — Audio player with speed control + questions below.
6. **Profile** — User settings, level selection.

## Implementation Plans

- [x] **Plan 1: Project Setup + Auth** — Docker, DB, Go project, auth endpoints, Next.js setup
  - `docs/agent-docs/plans/2026-04-02-project-setup-auth.md`
- [x] **Plan 2: Vocabulary (Backend + Frontend)** — Topics, words, SRS, flashcards, quiz
  - `docs/agent-docs/plans/2026-04-03-vocabulary-backend-frontend.md`
- [ ] **Plan 3: Listening (Backend + Frontend)** — Lessons, questions, audio player
- [ ] **Plan 4: Progress & Dashboard** — Progress endpoints, dashboard UI

## Future Additions (Post-MVP)

- AI conversation practice (Claude API chatbot)
- AI-generated content and personalized feedback
- Mobile app (React Native, reusing Go API)
- Pronunciation practice (speech recognition)
- Gamification (leaderboard, achievements)
- Social features (study groups)
