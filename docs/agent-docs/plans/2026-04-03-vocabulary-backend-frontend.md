# Vocabulary (Backend + Frontend) Implementation Plan

> Steps use checkbox (`- [ ]`) syntax for tracking progress.

**Goal:** Implement the full vocabulary feature — topics, words, flashcard learning, SRS review (SM-2), and quiz — across all DDD layers and the Next.js frontend.

**Architecture:** Follows the same DDD/Clean Architecture as auth: `domain/vocab/` (entities, SRS algorithm, repository interface) → `appcore/vocab/` (use cases, DTOs) → `driven/vocab/` (GORM repository) → `driving/httpui/` (vocab handler, presenter). Frontend pages live under `/dashboard/` to reuse the auth-guarded layout. Quiz questions are generated dynamically from words (no quiz table).

**Tech Stack:** Go, Gin, Uber fx, GORM, PostgreSQL, Next.js (App Router), React, Tailwind CSS.

> **Next.js 16 warning:** `web/AGENTS.md` states this version has breaking changes. Before writing any frontend code, read the relevant guide in `node_modules/next/dist/docs/` to verify API compatibility.

---

## Package Name Conventions

| Path | `package` name | Imported as |
|---|---|---|
| `internal/domain/vocab/` | `domain` | `vocabdom` when multiple domain pkgs in scope |
| `internal/appcore/vocab/` | `appcore` | `vocabcore` |
| `internal/driven/vocab/` | `driven` | `vocabdriven` |

---

## File Map

**Create:**
```
server/
├── migrations/
│   ├── 002_create_topics.up.sql
│   ├── 002_create_topics.down.sql
│   ├── 003_create_words.up.sql
│   ├── 003_create_words.down.sql
│   ├── 004_create_user_word_progress.up.sql
│   └── 004_create_user_word_progress.down.sql
├── internal/
│   ├── domain/vocab/
│   │   ├── entity.go
│   │   ├── errors.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── module.go
│   ├── appcore/vocab/
│   │   ├── dto.go
│   │   ├── usecase.go
│   │   ├── usecase_test.go
│   │   └── module.go
│   ├── driven/vocab/
│   │   ├── gorm_model.go
│   │   ├── repository.go
│   │   └── module.go
│   └── driving/httpui/
│       ├── handler/vocab_handler.go
│       ├── handler/vocab_handler_test.go
│       └── presenter/vocab_presenter.go
web/
├── lib/
│   └── vocab-api.ts
├── app/dashboard/
│   ├── topics/
│   │   ├── page.tsx
│   │   └── [id]/
│   │       ├── page.tsx
│   │       ├── learn/page.tsx
│   │       └── quiz/page.tsx
│   └── review/
│       └── page.tsx
└── components/
    ├── flashcard.tsx
    └── quiz-question.tsx
```

**Modify:**
```
server/internal/driving/httpui/server.go    — add vocab routes to protected group
server/internal/driving/httpui/module.go    — provide VocabHandler + VocabPresenter
server/cmd/api/main.go                     — wire vocab domain/appcore/driven modules
web/app/dashboard/layout.tsx               — add nav links (Chủ đề, Ôn tập)
web/app/dashboard/page.tsx                 — add vocab quick-links
```

---

## Task 1: Database Migrations

**Files:**
- Create: `server/migrations/002_create_topics.up.sql`
- Create: `server/migrations/002_create_topics.down.sql`
- Create: `server/migrations/003_create_words.up.sql`
- Create: `server/migrations/003_create_words.down.sql`
- Create: `server/migrations/004_create_user_word_progress.up.sql`
- Create: `server/migrations/004_create_user_word_progress.down.sql`

- [x] **Step 1: Create topics migration**

`server/migrations/002_create_topics.up.sql`:
```sql
CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    name_vi VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    level user_level NOT NULL DEFAULT 'beginner'
);

CREATE INDEX idx_topics_level ON topics(level);
```

`server/migrations/002_create_topics.down.sql`:
```sql
DROP TABLE IF EXISTS topics;
```

- [x] **Step 2: Create words migration**

`server/migrations/003_create_words.up.sql`:
```sql
CREATE TABLE words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(200) NOT NULL,
    ipa_phonetic VARCHAR(200) NOT NULL DEFAULT '',
    part_of_speech VARCHAR(50) NOT NULL DEFAULT '',
    vi_meaning VARCHAR(500) NOT NULL,
    en_definition TEXT NOT NULL DEFAULT '',
    example_sentence TEXT NOT NULL DEFAULT '',
    example_vi_translation TEXT NOT NULL DEFAULT '',
    audio_url TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    level user_level NOT NULL DEFAULT 'beginner',
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE
);

CREATE INDEX idx_words_topic_id ON words(topic_id);
CREATE INDEX idx_words_level ON words(level);
```

`server/migrations/003_create_words.down.sql`:
```sql
DROP TABLE IF EXISTS words;
```

- [x] **Step 3: Create user_word_progress migration**

`server/migrations/004_create_user_word_progress.up.sql`:
```sql
CREATE TABLE user_word_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word_id UUID NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    ease_factor DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    interval_days INTEGER NOT NULL DEFAULT 0,
    next_review_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    review_count INTEGER NOT NULL DEFAULT 0,
    last_reviewed_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, word_id)
);

CREATE INDEX idx_user_word_progress_due ON user_word_progress(user_id, next_review_at);
```

`server/migrations/004_create_user_word_progress.down.sql`:
```sql
DROP TABLE IF EXISTS user_word_progress;
```

- [x] **Step 4: Apply migrations**

```bash
cd server && docker-compose up -d postgres
psql "$DATABASE_URL" -f migrations/002_create_topics.up.sql
psql "$DATABASE_URL" -f migrations/003_create_words.up.sql
psql "$DATABASE_URL" -f migrations/004_create_user_word_progress.up.sql
```

Expected: All three tables created without errors. Verify with `\dt` in psql.

---

## Task 2: Domain Layer — Entities, Errors, Repository Interface

**Files:**
- Create: `server/internal/domain/vocab/entity.go`
- Create: `server/internal/domain/vocab/errors.go`
- Create: `server/internal/domain/vocab/repository.go`

- [x] **Step 1: Implement entity.go**

`server/internal/domain/vocab/entity.go`:
```go
package domain

import "time"

type Topic struct {
	ID          string
	Name        string
	NameVI      string
	Description string
	Level       string
}

type Word struct {
	ID                   string
	Word                 string
	IPAPhonetic          string
	PartOfSpeech         string
	VIMeaning            string
	ENDefinition         string
	ExampleSentence      string
	ExampleVITranslation string
	AudioURL             string
	ImageURL             string
	Level                string
	TopicID              string
}

type UserWordProgress struct {
	UserID         string
	WordID         string
	EaseFactor     float64
	IntervalDays   int
	NextReviewAt   time.Time
	ReviewCount    int
	LastReviewedAt *time.Time
}
```

- [x] **Step 2: Implement errors.go**

`server/internal/domain/vocab/errors.go`:
```go
package domain

import "errors"

var (
	ErrTopicNotFound = errors.New("topic not found")
	ErrWordNotFound  = errors.New("word not found")
)
```

- [x] **Step 3: Implement repository.go**

`server/internal/domain/vocab/repository.go`:
```go
package domain

import (
	"context"
	"time"
)

type Repository interface {
	// Topics
	ListTopics(ctx context.Context, level string) ([]Topic, error)
	GetTopicByID(ctx context.Context, id string) (*Topic, error)

	// Words
	ListWordsByTopic(ctx context.Context, topicID string) ([]Word, error)
	GetWordByID(ctx context.Context, id string) (*Word, error)
	GetRandomWords(ctx context.Context, topicID string, excludeID string, limit int) ([]Word, error)

	// Progress
	GetProgress(ctx context.Context, userID, wordID string) (*UserWordProgress, error)
	UpsertProgress(ctx context.Context, progress *UserWordProgress) error
	GetDueWords(ctx context.Context, userID string, now time.Time, limit int) ([]Word, error)
}
```

- [x] **Step 4: Verify it compiles**

```bash
cd server && go build ./internal/domain/vocab/...
```

Expected: Exit 0, no errors.

---

## Task 3: Domain SRS Service (TDD)

**Files:**
- Create: `server/internal/domain/vocab/service_test.go`
- Create: `server/internal/domain/vocab/service.go`
- Create: `server/internal/domain/vocab/module.go`

- [x] **Step 1: Write the failing tests**

`server/internal/domain/vocab/service_test.go`:
```go
package domain_test

import (
	"math"
	"testing"
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

func TestCalculateReview_FirstReview_Easy(t *testing.T) {
	svc := domain.NewService()
	now := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	progress := &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "word-1",
		EaseFactor:   2.5,
		IntervalDays: 0,
		ReviewCount:  0,
	}

	result := svc.CalculateReview(progress, 5, now)

	if result.IntervalDays != 1 {
		t.Errorf("IntervalDays = %d, want 1", result.IntervalDays)
	}
	// EF = 2.5 + (0.1 - (5-5)*(0.08+(5-5)*0.02)) = 2.5 + 0.1 = 2.6
	if math.Abs(result.EaseFactor-2.6) > 0.001 {
		t.Errorf("EaseFactor = %f, want 2.6", result.EaseFactor)
	}
	expected := now.Add(24 * time.Hour)
	if !result.NextReviewAt.Equal(expected) {
		t.Errorf("NextReviewAt = %v, want %v", result.NextReviewAt, expected)
	}
}

func TestCalculateReview_SecondReview_OK(t *testing.T) {
	svc := domain.NewService()
	now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	progress := &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "word-1",
		EaseFactor:   2.6,
		IntervalDays: 1,
		ReviewCount:  1,
	}

	result := svc.CalculateReview(progress, 3, now)

	if result.IntervalDays != 6 {
		t.Errorf("IntervalDays = %d, want 6", result.IntervalDays)
	}
	// EF = 2.6 + (0.1 - (5-3)*(0.08+(5-3)*0.02)) = 2.6 + (0.1 - 2*0.12) = 2.6 - 0.14 = 2.46
	if math.Abs(result.EaseFactor-2.46) > 0.001 {
		t.Errorf("EaseFactor = %f, want 2.46", result.EaseFactor)
	}
	expected := now.Add(6 * 24 * time.Hour)
	if !result.NextReviewAt.Equal(expected) {
		t.Errorf("NextReviewAt = %v, want %v", result.NextReviewAt, expected)
	}
}

func TestCalculateReview_ThirdReview_Easy(t *testing.T) {
	svc := domain.NewService()
	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	progress := &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "word-1",
		EaseFactor:   2.46,
		IntervalDays: 6,
		ReviewCount:  2,
	}

	result := svc.CalculateReview(progress, 5, now)

	// interval = round(6 * 2.46) = round(14.76) = 15
	if result.IntervalDays != 15 {
		t.Errorf("IntervalDays = %d, want 15", result.IntervalDays)
	}
}

func TestCalculateReview_Hard_ResetsInterval(t *testing.T) {
	svc := domain.NewService()
	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	progress := &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "word-1",
		EaseFactor:   2.5,
		IntervalDays: 15,
		ReviewCount:  3,
	}

	result := svc.CalculateReview(progress, 1, now)

	if result.IntervalDays != 1 {
		t.Errorf("IntervalDays = %d, want 1 (reset)", result.IntervalDays)
	}
	// EF = 2.5 + (0.1 - (5-1)*(0.08+(5-1)*0.02)) = 2.5 + (0.1 - 4*0.16) = 2.5 - 0.54 = 1.96
	if math.Abs(result.EaseFactor-1.96) > 0.001 {
		t.Errorf("EaseFactor = %f, want 1.96", result.EaseFactor)
	}
	expected := now.Add(24 * time.Hour)
	if !result.NextReviewAt.Equal(expected) {
		t.Errorf("NextReviewAt = %v, want %v", result.NextReviewAt, expected)
	}
}

func TestCalculateReview_EaseFactorFloor(t *testing.T) {
	svc := domain.NewService()
	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	progress := &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "word-1",
		EaseFactor:   1.3,
		IntervalDays: 1,
		ReviewCount:  1,
	}

	result := svc.CalculateReview(progress, 1, now)

	if result.EaseFactor < 1.3 {
		t.Errorf("EaseFactor = %f, should not go below 1.3", result.EaseFactor)
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./internal/domain/vocab/... -v
```

Expected: `FAIL` — `domain.NewService` undefined.

- [x] **Step 3: Implement service.go**

`server/internal/domain/vocab/service.go`:
```go
package domain

import (
	"math"
	"time"
)

type Service struct{}

func NewService() *Service { return &Service{} }

type ReviewResult struct {
	EaseFactor   float64
	IntervalDays int
	NextReviewAt time.Time
}

// CalculateReview applies the SM-2 algorithm.
// quality: 1 (Hard), 3 (OK), 5 (Easy).
func (s *Service) CalculateReview(current *UserWordProgress, quality int, now time.Time) ReviewResult {
	ef := current.EaseFactor
	if ef < 1.3 {
		ef = 2.5
	}

	var interval int
	if quality < 3 {
		interval = 1
	} else {
		switch current.ReviewCount {
		case 0:
			interval = 1
		case 1:
			interval = 6
		default:
			interval = int(math.Round(float64(current.IntervalDays) * ef))
		}
	}

	q := float64(quality)
	ef = ef + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if ef < 1.3 {
		ef = 1.3
	}

	return ReviewResult{
		EaseFactor:   ef,
		IntervalDays: interval,
		NextReviewAt: now.Add(time.Duration(interval) * 24 * time.Hour),
	}
}
```

- [x] **Step 4: Implement module.go**

`server/internal/domain/vocab/module.go`:
```go
package domain

import "go.uber.org/fx"

var Module = fx.Module("vocab-domain",
	fx.Provide(NewService),
)
```

- [x] **Step 5: Run tests to verify they pass**

```bash
cd server && go test ./internal/domain/vocab/... -v
```

Expected: `PASS` for all five test functions.

---

## Task 4: Appcore DTOs

**Files:**
- Create: `server/internal/appcore/vocab/dto.go`

- [x] **Step 1: Implement dto.go**

`server/internal/appcore/vocab/dto.go`:
```go
package appcore

type TopicOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameVI      string `json:"name_vi"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

type WordOutput struct {
	ID                   string `json:"id"`
	Word                 string `json:"word"`
	IPAPhonetic          string `json:"ipa_phonetic"`
	PartOfSpeech         string `json:"part_of_speech"`
	VIMeaning            string `json:"vi_meaning"`
	ENDefinition         string `json:"en_definition"`
	ExampleSentence      string `json:"example_sentence"`
	ExampleVITranslation string `json:"example_vi_translation"`
	AudioURL             string `json:"audio_url"`
	ImageURL             string `json:"image_url"`
	Level                string `json:"level"`
	TopicID              string `json:"topic_id"`
}

type ReviewInput struct {
	Quality int `json:"quality"` // 1 (Hard), 3 (OK), 5 (Easy)
}

type QuizQuestion struct {
	WordID  string   `json:"word_id"`
	Word    string   `json:"word"`
	Options []string `json:"options"`
}

type QuizAnswerInput struct {
	Answers []QuizAnswer `json:"answers"`
}

type QuizAnswer struct {
	WordID string `json:"word_id"`
	Answer string `json:"answer"`
}

type QuizResult struct {
	Score   int              `json:"score"`
	Total   int              `json:"total"`
	Results []QuizItemResult `json:"results"`
}

type QuizItemResult struct {
	WordID        string `json:"word_id"`
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correct_answer"`
}
```

---

## Task 5: Appcore Use Cases (TDD)

**Files:**
- Create: `server/internal/appcore/vocab/usecase_test.go`
- Create: `server/internal/appcore/vocab/usecase.go`
- Create: `server/internal/appcore/vocab/module.go`

- [x] **Step 1: Write the failing tests**

`server/internal/appcore/vocab/usecase_test.go`:
```go
package appcore_test

import (
	"context"
	"testing"
	"time"

	appcore "github.com/sonpham/vielish/server/internal/appcore/vocab"
	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

// --- Stub Repository ---

type stubRepo struct {
	topics   []domain.Topic
	words    []domain.Word
	progress map[string]*domain.UserWordProgress // key: "userID:wordID"
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		progress: make(map[string]*domain.UserWordProgress),
	}
}

func (r *stubRepo) ListTopics(_ context.Context, level string) ([]domain.Topic, error) {
	if level == "" {
		return r.topics, nil
	}
	var out []domain.Topic
	for _, t := range r.topics {
		if t.Level == level {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *stubRepo) GetTopicByID(_ context.Context, id string) (*domain.Topic, error) {
	for _, t := range r.topics {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, domain.ErrTopicNotFound
}

func (r *stubRepo) ListWordsByTopic(_ context.Context, topicID string) ([]domain.Word, error) {
	var out []domain.Word
	for _, w := range r.words {
		if w.TopicID == topicID {
			out = append(out, w)
		}
	}
	return out, nil
}

func (r *stubRepo) GetWordByID(_ context.Context, id string) (*domain.Word, error) {
	for _, w := range r.words {
		if w.ID == id {
			return &w, nil
		}
	}
	return nil, domain.ErrWordNotFound
}

func (r *stubRepo) GetRandomWords(_ context.Context, topicID, excludeID string, limit int) ([]domain.Word, error) {
	var out []domain.Word
	for _, w := range r.words {
		if w.TopicID == topicID && w.ID != excludeID && len(out) < limit {
			out = append(out, w)
		}
	}
	return out, nil
}

func (r *stubRepo) GetProgress(_ context.Context, userID, wordID string) (*domain.UserWordProgress, error) {
	key := userID + ":" + wordID
	if p, ok := r.progress[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (r *stubRepo) UpsertProgress(_ context.Context, p *domain.UserWordProgress) error {
	key := p.UserID + ":" + p.WordID
	r.progress[key] = p
	return nil
}

func (r *stubRepo) GetDueWords(_ context.Context, userID string, now time.Time, limit int) ([]domain.Word, error) {
	var out []domain.Word
	for key, p := range r.progress {
		_ = key
		if p.UserID == userID && !p.NextReviewAt.After(now) && len(out) < limit {
			for _, w := range r.words {
				if w.ID == p.WordID {
					out = append(out, w)
					break
				}
			}
		}
	}
	return out, nil
}

// --- Tests ---

func TestListTopics(t *testing.T) {
	repo := newStubRepo()
	repo.topics = []domain.Topic{
		{ID: "t1", Name: "Food", NameVI: "Thức ăn", Level: "beginner"},
		{ID: "t2", Name: "Travel", NameVI: "Du lịch", Level: "intermediate"},
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	out, err := uc.ListTopics(context.Background(), "beginner")
	if err != nil {
		t.Fatalf("ListTopics error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].ID != "t1" {
		t.Errorf("ID = %q, want %q", out[0].ID, "t1")
	}
}

func TestGetTopicWords(t *testing.T) {
	repo := newStubRepo()
	repo.topics = []domain.Topic{{ID: "t1", Name: "Food", NameVI: "Thức ăn", Level: "beginner"}}
	repo.words = []domain.Word{
		{ID: "w1", Word: "apple", VIMeaning: "quả táo", TopicID: "t1"},
		{ID: "w2", Word: "rice", VIMeaning: "cơm", TopicID: "t1"},
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	out, err := uc.GetTopicWords(context.Background(), "t1")
	if err != nil {
		t.Fatalf("GetTopicWords error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
}

func TestSubmitReview_CreatesProgress(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{{ID: "w1", Word: "apple", TopicID: "t1"}}
	uc := appcore.NewUseCase(repo, domain.NewService())

	err := uc.SubmitReview(context.Background(), "user-1", "w1", appcore.ReviewInput{Quality: 5})
	if err != nil {
		t.Fatalf("SubmitReview error: %v", err)
	}

	p, _ := repo.GetProgress(context.Background(), "user-1", "w1")
	if p == nil {
		t.Fatal("expected progress to be created")
	}
	if p.ReviewCount != 1 {
		t.Errorf("ReviewCount = %d, want 1", p.ReviewCount)
	}
	if p.IntervalDays != 1 {
		t.Errorf("IntervalDays = %d, want 1", p.IntervalDays)
	}
}

func TestSubmitReview_InvalidQuality(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{{ID: "w1", Word: "apple", TopicID: "t1"}}
	uc := appcore.NewUseCase(repo, domain.NewService())

	err := uc.SubmitReview(context.Background(), "user-1", "w1", appcore.ReviewInput{Quality: 4})
	if err == nil {
		t.Fatal("expected error for invalid quality")
	}
}

func TestGetQuiz(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{
		{ID: "w1", Word: "apple", VIMeaning: "quả táo", TopicID: "t1"},
		{ID: "w2", Word: "rice", VIMeaning: "cơm", TopicID: "t1"},
		{ID: "w3", Word: "water", VIMeaning: "nước", TopicID: "t1"},
		{ID: "w4", Word: "bread", VIMeaning: "bánh mì", TopicID: "t1"},
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	questions, err := uc.GetQuiz(context.Background(), "t1")
	if err != nil {
		t.Fatalf("GetQuiz error: %v", err)
	}
	if len(questions) != 4 {
		t.Fatalf("len = %d, want 4", len(questions))
	}
	for _, q := range questions {
		if len(q.Options) != 4 {
			t.Errorf("word %q: options count = %d, want 4", q.Word, len(q.Options))
		}
	}
}

func TestSubmitQuiz(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{
		{ID: "w1", Word: "apple", VIMeaning: "quả táo", TopicID: "t1"},
		{ID: "w2", Word: "rice", VIMeaning: "cơm", TopicID: "t1"},
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	result, err := uc.SubmitQuiz(context.Background(), "user-1", "t1", appcore.QuizAnswerInput{
		Answers: []appcore.QuizAnswer{
			{WordID: "w1", Answer: "quả táo"},
			{WordID: "w2", Answer: "wrong"},
		},
	})
	if err != nil {
		t.Fatalf("SubmitQuiz error: %v", err)
	}
	if result.Score != 1 {
		t.Errorf("Score = %d, want 1", result.Score)
	}
	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./internal/appcore/vocab/... -v
```

Expected: `FAIL` — `appcore.NewUseCase` undefined.

- [x] **Step 3: Implement usecase.go**

`server/internal/appcore/vocab/usecase.go`:
```go
package appcore

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

var validQualities = map[int]bool{1: true, 3: true, 5: true}

type UseCase struct {
	repo    domain.Repository
	service *domain.Service
}

func NewUseCase(repo domain.Repository, service *domain.Service) *UseCase {
	return &UseCase{repo: repo, service: service}
}

func (uc *UseCase) ListTopics(ctx context.Context, level string) ([]TopicOutput, error) {
	topics, err := uc.repo.ListTopics(ctx, level)
	if err != nil {
		return nil, err
	}
	out := make([]TopicOutput, len(topics))
	for i, t := range topics {
		out[i] = TopicOutput{
			ID:          t.ID,
			Name:        t.Name,
			NameVI:      t.NameVI,
			Description: t.Description,
			Level:       t.Level,
		}
	}
	return out, nil
}

func (uc *UseCase) GetTopicWords(ctx context.Context, topicID string) ([]WordOutput, error) {
	words, err := uc.repo.ListWordsByTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}
	return toWordOutputs(words), nil
}

func (uc *UseCase) GetWord(ctx context.Context, id string) (*WordOutput, error) {
	w, err := uc.repo.GetWordByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toWordOutput(*w)
	return &out, nil
}

func (uc *UseCase) GetDueReviews(ctx context.Context, userID string) ([]WordOutput, error) {
	words, err := uc.repo.GetDueWords(ctx, userID, time.Now(), 20)
	if err != nil {
		return nil, err
	}
	return toWordOutputs(words), nil
}

func (uc *UseCase) SubmitReview(ctx context.Context, userID, wordID string, input ReviewInput) error {
	if !validQualities[input.Quality] {
		return fmt.Errorf("invalid quality %d: must be 1, 3, or 5", input.Quality)
	}

	if _, err := uc.repo.GetWordByID(ctx, wordID); err != nil {
		return err
	}

	progress, err := uc.repo.GetProgress(ctx, userID, wordID)
	if err != nil {
		return err
	}
	if progress == nil {
		progress = &domain.UserWordProgress{
			UserID:     userID,
			WordID:     wordID,
			EaseFactor: 2.5,
		}
	}

	now := time.Now()
	result := uc.service.CalculateReview(progress, input.Quality, now)

	progress.EaseFactor = result.EaseFactor
	progress.IntervalDays = result.IntervalDays
	progress.NextReviewAt = result.NextReviewAt
	progress.ReviewCount++
	progress.LastReviewedAt = &now

	return uc.repo.UpsertProgress(ctx, progress)
}

func (uc *UseCase) GetQuiz(ctx context.Context, topicID string) ([]QuizQuestion, error) {
	words, err := uc.repo.ListWordsByTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	questions := make([]QuizQuestion, len(words))
	for i, w := range words {
		distractors, err := uc.repo.GetRandomWords(ctx, topicID, w.ID, 3)
		if err != nil {
			return nil, err
		}

		options := make([]string, 0, 4)
		options = append(options, w.VIMeaning)
		for _, d := range distractors {
			options = append(options, d.VIMeaning)
		}
		rand.Shuffle(len(options), func(i, j int) {
			options[i], options[j] = options[j], options[i]
		})

		questions[i] = QuizQuestion{
			WordID:  w.ID,
			Word:    w.Word,
			Options: options,
		}
	}
	return questions, nil
}

func (uc *UseCase) SubmitQuiz(ctx context.Context, userID, topicID string, input QuizAnswerInput) (*QuizResult, error) {
	answerMap := make(map[string]string, len(input.Answers))
	for _, a := range input.Answers {
		answerMap[a.WordID] = a.Answer
	}

	words, err := uc.repo.ListWordsByTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	wordMap := make(map[string]domain.Word, len(words))
	for _, w := range words {
		wordMap[w.ID] = w
	}

	var score int
	results := make([]QuizItemResult, 0, len(input.Answers))
	for _, a := range input.Answers {
		w, ok := wordMap[a.WordID]
		if !ok {
			continue
		}
		correct := a.Answer == w.VIMeaning
		if correct {
			score++
		}
		results = append(results, QuizItemResult{
			WordID:        a.WordID,
			Correct:       correct,
			CorrectAnswer: w.VIMeaning,
		})
	}

	return &QuizResult{
		Score:   score,
		Total:   len(results),
		Results: results,
	}, nil
}

func toWordOutput(w domain.Word) WordOutput {
	return WordOutput{
		ID:                   w.ID,
		Word:                 w.Word,
		IPAPhonetic:          w.IPAPhonetic,
		PartOfSpeech:         w.PartOfSpeech,
		VIMeaning:            w.VIMeaning,
		ENDefinition:         w.ENDefinition,
		ExampleSentence:      w.ExampleSentence,
		ExampleVITranslation: w.ExampleVITranslation,
		AudioURL:             w.AudioURL,
		ImageURL:             w.ImageURL,
		Level:                w.Level,
		TopicID:              w.TopicID,
	}
}

func toWordOutputs(words []domain.Word) []WordOutput {
	out := make([]WordOutput, len(words))
	for i, w := range words {
		out[i] = toWordOutput(w)
	}
	return out
}
```

- [x] **Step 4: Implement module.go**

`server/internal/appcore/vocab/module.go`:
```go
package appcore

import "go.uber.org/fx"

var Module = fx.Module("vocab-appcore",
	fx.Provide(NewUseCase),
)
```

- [x] **Step 5: Run tests to verify they pass**

```bash
cd server && go test ./internal/appcore/vocab/... -v
```

Expected: `PASS` for all test functions.

---

## Task 6: Driven GORM Models + Repository

**Files:**
- Create: `server/internal/driven/vocab/gorm_model.go`
- Create: `server/internal/driven/vocab/repository.go`
- Create: `server/internal/driven/vocab/module.go`

- [x] **Step 1: Implement gorm_model.go**

`server/internal/driven/vocab/gorm_model.go`:
```go
package driven

import (
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

type TopicModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string `gorm:"not null"`
	NameVI      string `gorm:"column:name_vi;not null"`
	Description string
	Level       string `gorm:"not null"`
}

func (TopicModel) TableName() string { return "topics" }

func (m *TopicModel) ToEntity() *domain.Topic {
	return &domain.Topic{
		ID:          m.ID,
		Name:        m.Name,
		NameVI:      m.NameVI,
		Description: m.Description,
		Level:       m.Level,
	}
}

type WordModel struct {
	ID                   string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Word                 string `gorm:"not null"`
	IPAPhonetic          string `gorm:"column:ipa_phonetic"`
	PartOfSpeech         string `gorm:"column:part_of_speech"`
	VIMeaning            string `gorm:"column:vi_meaning;not null"`
	ENDefinition         string `gorm:"column:en_definition"`
	ExampleSentence      string `gorm:"column:example_sentence"`
	ExampleVITranslation string `gorm:"column:example_vi_translation"`
	AudioURL             string `gorm:"column:audio_url"`
	ImageURL             string `gorm:"column:image_url"`
	Level                string `gorm:"not null"`
	TopicID              string `gorm:"column:topic_id;type:uuid;not null"`
}

func (WordModel) TableName() string { return "words" }

func (m *WordModel) ToEntity() *domain.Word {
	return &domain.Word{
		ID:                   m.ID,
		Word:                 m.Word,
		IPAPhonetic:          m.IPAPhonetic,
		PartOfSpeech:         m.PartOfSpeech,
		VIMeaning:            m.VIMeaning,
		ENDefinition:         m.ENDefinition,
		ExampleSentence:      m.ExampleSentence,
		ExampleVITranslation: m.ExampleVITranslation,
		AudioURL:             m.AudioURL,
		ImageURL:             m.ImageURL,
		Level:                m.Level,
		TopicID:              m.TopicID,
	}
}

type UserWordProgressModel struct {
	UserID         string     `gorm:"primaryKey;type:uuid"`
	WordID         string     `gorm:"primaryKey;type:uuid"`
	EaseFactor     float64    `gorm:"column:ease_factor;default:2.5"`
	IntervalDays   int        `gorm:"column:interval_days;default:0"`
	NextReviewAt   time.Time  `gorm:"column:next_review_at"`
	ReviewCount    int        `gorm:"column:review_count;default:0"`
	LastReviewedAt *time.Time `gorm:"column:last_reviewed_at"`
}

func (UserWordProgressModel) TableName() string { return "user_word_progress" }

func (m *UserWordProgressModel) ToEntity() *domain.UserWordProgress {
	return &domain.UserWordProgress{
		UserID:         m.UserID,
		WordID:         m.WordID,
		EaseFactor:     m.EaseFactor,
		IntervalDays:   m.IntervalDays,
		NextReviewAt:   m.NextReviewAt,
		ReviewCount:    m.ReviewCount,
		LastReviewedAt: m.LastReviewedAt,
	}
}
```

- [x] **Step 2: Implement repository.go**

`server/internal/driven/vocab/repository.go`:
```go
package driven

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListTopics(ctx context.Context, level string) ([]domain.Topic, error) {
	var models []TopicModel
	q := r.db.WithContext(ctx)
	if level != "" {
		q = q.Where("level = ?", level)
	}
	if err := q.Order("name ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("listing topics: %w", err)
	}
	topics := make([]domain.Topic, len(models))
	for i, m := range models {
		topics[i] = *m.ToEntity()
	}
	return topics, nil
}

func (r *Repository) GetTopicByID(ctx context.Context, id string) (*domain.Topic, error) {
	var m TopicModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTopicNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting topic: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) ListWordsByTopic(ctx context.Context, topicID string) ([]domain.Word, error) {
	var models []WordModel
	err := r.db.WithContext(ctx).Where("topic_id = ?", topicID).Order("word ASC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("listing words: %w", err)
	}
	words := make([]domain.Word, len(models))
	for i, m := range models {
		words[i] = *m.ToEntity()
	}
	return words, nil
}

func (r *Repository) GetWordByID(ctx context.Context, id string) (*domain.Word, error) {
	var m WordModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrWordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting word: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) GetRandomWords(ctx context.Context, topicID, excludeID string, limit int) ([]domain.Word, error) {
	var models []WordModel
	err := r.db.WithContext(ctx).
		Where("topic_id = ? AND id != ?", topicID, excludeID).
		Order("RANDOM()").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("getting random words: %w", err)
	}
	words := make([]domain.Word, len(models))
	for i, m := range models {
		words[i] = *m.ToEntity()
	}
	return words, nil
}

func (r *Repository) GetProgress(ctx context.Context, userID, wordID string) (*domain.UserWordProgress, error) {
	var m UserWordProgressModel
	err := r.db.WithContext(ctx).Where("user_id = ? AND word_id = ?", userID, wordID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting progress: %w", err)
	}
	return m.ToEntity(), nil
}

func (r *Repository) UpsertProgress(ctx context.Context, p *domain.UserWordProgress) error {
	m := &UserWordProgressModel{
		UserID:         p.UserID,
		WordID:         p.WordID,
		EaseFactor:     p.EaseFactor,
		IntervalDays:   p.IntervalDays,
		NextReviewAt:   p.NextReviewAt,
		ReviewCount:    p.ReviewCount,
		LastReviewedAt: p.LastReviewedAt,
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "word_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"ease_factor", "interval_days", "next_review_at", "review_count", "last_reviewed_at"}),
		}).
		Create(m).Error
	if err != nil {
		return fmt.Errorf("upserting progress: %w", err)
	}
	return nil
}

func (r *Repository) GetDueWords(ctx context.Context, userID string, now time.Time, limit int) ([]domain.Word, error) {
	var models []WordModel
	err := r.db.WithContext(ctx).
		Joins("JOIN user_word_progress p ON words.id = p.word_id").
		Where("p.user_id = ? AND p.next_review_at <= ?", userID, now).
		Order("p.next_review_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("getting due words: %w", err)
	}
	words := make([]domain.Word, len(models))
	for i, m := range models {
		words[i] = *m.ToEntity()
	}
	return words, nil
}
```

- [x] **Step 3: Implement module.go**

`server/internal/driven/vocab/module.go`:
```go
package driven

import (
	"go.uber.org/fx"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

var Module = fx.Module("vocab-driven",
	fx.Provide(
		fx.Annotate(
			NewRepository,
			fx.As(new(domain.Repository)),
		),
	),
)
```

- [x] **Step 4: Verify it compiles**

```bash
cd server && go build ./internal/driven/vocab/...
```

Expected: Exit 0, no errors.

---

## Task 7: HTTP Vocab Handler (TDD)

**Files:**
- Create: `server/internal/driving/httpui/handler/vocab_handler_test.go`
- Create: `server/internal/driving/httpui/handler/vocab_handler.go`
- Create: `server/internal/driving/httpui/presenter/vocab_presenter.go`

- [x] **Step 1: Write the failing tests**

`server/internal/driving/httpui/handler/vocab_handler_test.go`:
```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/vocab"
	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
	"github.com/sonpham/vielish/server/pkg/ctxbase"
)

func init() { gin.SetMode(gin.TestMode) }

// --- Stub ---

type stubVocabUC struct {
	topics    []appcore.TopicOutput
	words     []appcore.WordOutput
	questions []appcore.QuizQuestion
	quiz      *appcore.QuizResult
	reviewErr error
}

func (s *stubVocabUC) ListTopics(_ context.Context, _ string) ([]appcore.TopicOutput, error) {
	return s.topics, nil
}
func (s *stubVocabUC) GetTopicWords(_ context.Context, _ string) ([]appcore.WordOutput, error) {
	return s.words, nil
}
func (s *stubVocabUC) GetWord(_ context.Context, _ string) (*appcore.WordOutput, error) {
	if len(s.words) > 0 {
		return &s.words[0], nil
	}
	return nil, nil
}
func (s *stubVocabUC) GetDueReviews(_ context.Context, _ string) ([]appcore.WordOutput, error) {
	return s.words, nil
}
func (s *stubVocabUC) SubmitReview(_ context.Context, _, _ string, _ appcore.ReviewInput) error {
	return s.reviewErr
}
func (s *stubVocabUC) GetQuiz(_ context.Context, _ string) ([]appcore.QuizQuestion, error) {
	return s.questions, nil
}
func (s *stubVocabUC) SubmitQuiz(_ context.Context, _, _ string, _ appcore.QuizAnswerInput) (*appcore.QuizResult, error) {
	return s.quiz, nil
}

func setUserID(c *gin.Context, userID string) {
	ctx := ctxbase.SetUserID(c.Request.Context(), userID)
	c.Request = c.Request.WithContext(ctx)
}

func TestVocabHandler_ListTopics(t *testing.T) {
	stub := &stubVocabUC{
		topics: []appcore.TopicOutput{
			{ID: "t1", Name: "Food", NameVI: "Thức ăn", Level: "beginner"},
		},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/topics?level=beginner", nil)

	h.ListTopics(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body []map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if len(body) != 1 {
		t.Fatalf("len = %d, want 1", len(body))
	}
	if body[0]["name"] != "Food" {
		t.Errorf("name = %v, want Food", body[0]["name"])
	}
}

func TestVocabHandler_SubmitReview(t *testing.T) {
	stub := &stubVocabUC{}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	body := `{"quality":5}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/review/w1", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "wordId", Value: "w1"}}
	setUserID(c, "user-1")

	h.SubmitReview(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestVocabHandler_GetQuiz(t *testing.T) {
	stub := &stubVocabUC{
		questions: []appcore.QuizQuestion{
			{WordID: "w1", Word: "apple", Options: []string{"quả táo", "cơm", "nước", "bánh mì"}},
		},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/quiz/t1", nil)
	c.Params = gin.Params{{Key: "topicId", Value: "t1"}}

	h.GetQuiz(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	questions, ok := body["questions"].([]any)
	if !ok || len(questions) != 1 {
		t.Fatalf("questions count = %v, want 1", body["questions"])
	}
}
```

- [x] **Step 2: Run to verify it fails**

```bash
cd server && go test ./internal/driving/httpui/handler/... -run TestVocab -v
```

Expected: `FAIL` — `handler.NewVocabHandlerFromInterface` undefined.

- [x] **Step 3: Implement vocab_presenter.go**

`server/internal/driving/httpui/presenter/vocab_presenter.go`:
```go
package presenter

import (
	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/vocab"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

type VocabPresenter struct{}

func NewVocabPresenter() *VocabPresenter { return &VocabPresenter{} }

func (p *VocabPresenter) Topics(c *gin.Context, status int, topics []appcore.TopicOutput) {
	httpbase.Success(c, status, topics)
}

func (p *VocabPresenter) Words(c *gin.Context, status int, words []appcore.WordOutput) {
	httpbase.Success(c, status, words)
}

func (p *VocabPresenter) Word(c *gin.Context, status int, word *appcore.WordOutput) {
	httpbase.Success(c, status, word)
}

func (p *VocabPresenter) Quiz(c *gin.Context, status int, questions []appcore.QuizQuestion) {
	httpbase.Success(c, status, gin.H{"questions": questions})
}

func (p *VocabPresenter) QuizResult(c *gin.Context, status int, result *appcore.QuizResult) {
	httpbase.Success(c, status, result)
}
```

- [x] **Step 4: Implement vocab_handler.go**

`server/internal/driving/httpui/handler/vocab_handler.go`:
```go
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	appcore "github.com/sonpham/vielish/server/internal/appcore/vocab"
	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
	"github.com/sonpham/vielish/server/pkg/ctxbase"
	"github.com/sonpham/vielish/server/pkg/httpbase"
)

type VocabUseCaseInterface interface {
	ListTopics(ctx context.Context, level string) ([]appcore.TopicOutput, error)
	GetTopicWords(ctx context.Context, topicID string) ([]appcore.WordOutput, error)
	GetWord(ctx context.Context, id string) (*appcore.WordOutput, error)
	GetDueReviews(ctx context.Context, userID string) ([]appcore.WordOutput, error)
	SubmitReview(ctx context.Context, userID, wordID string, input appcore.ReviewInput) error
	GetQuiz(ctx context.Context, topicID string) ([]appcore.QuizQuestion, error)
	SubmitQuiz(ctx context.Context, userID, topicID string, input appcore.QuizAnswerInput) (*appcore.QuizResult, error)
}

type VocabHandler struct {
	useCase   VocabUseCaseInterface
	presenter *presenter.VocabPresenter
}

func NewVocabHandler(uc *appcore.UseCase, p *presenter.VocabPresenter) *VocabHandler {
	return &VocabHandler{useCase: uc, presenter: p}
}

func NewVocabHandlerFromInterface(uc VocabUseCaseInterface, p *presenter.VocabPresenter) *VocabHandler {
	return &VocabHandler{useCase: uc, presenter: p}
}

func (h *VocabHandler) ListTopics(c *gin.Context) {
	level := c.Query("level")
	topics, err := h.useCase.ListTopics(c.Request.Context(), level)
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to list topics")
		return
	}
	h.presenter.Topics(c, http.StatusOK, topics)
}

func (h *VocabHandler) GetTopicWords(c *gin.Context) {
	topicID := c.Param("id")
	words, err := h.useCase.GetTopicWords(c.Request.Context(), topicID)
	if errors.Is(err, domain.ErrTopicNotFound) {
		httpbase.Error(c, http.StatusNotFound, "topic not found")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to get words")
		return
	}
	h.presenter.Words(c, http.StatusOK, words)
}

func (h *VocabHandler) GetWord(c *gin.Context) {
	id := c.Param("id")
	word, err := h.useCase.GetWord(c.Request.Context(), id)
	if errors.Is(err, domain.ErrWordNotFound) {
		httpbase.Error(c, http.StatusNotFound, "word not found")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to get word")
		return
	}
	h.presenter.Word(c, http.StatusOK, word)
}

func (h *VocabHandler) GetDueReviews(c *gin.Context) {
	userID, ok := ctxbase.GetUserID(c.Request.Context())
	if !ok {
		httpbase.Error(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	words, err := h.useCase.GetDueReviews(c.Request.Context(), userID)
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to get due reviews")
		return
	}
	h.presenter.Words(c, http.StatusOK, words)
}

type reviewRequest struct {
	Quality int `json:"quality" binding:"required"`
}

func (h *VocabHandler) SubmitReview(c *gin.Context) {
	userID, ok := ctxbase.GetUserID(c.Request.Context())
	if !ok {
		httpbase.Error(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	wordID := c.Param("wordId")

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	err := h.useCase.SubmitReview(c.Request.Context(), userID, wordID, appcore.ReviewInput{Quality: req.Quality})
	if errors.Is(err, domain.ErrWordNotFound) {
		httpbase.Error(c, http.StatusNotFound, "word not found")
		return
	}
	if err != nil {
		httpbase.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	httpbase.Success(c, http.StatusOK, gin.H{"message": "review submitted"})
}

func (h *VocabHandler) GetQuiz(c *gin.Context) {
	topicID := c.Param("topicId")
	questions, err := h.useCase.GetQuiz(c.Request.Context(), topicID)
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to generate quiz")
		return
	}
	h.presenter.Quiz(c, http.StatusOK, questions)
}

type quizSubmitRequest struct {
	Answers []appcore.QuizAnswer `json:"answers" binding:"required"`
}

func (h *VocabHandler) SubmitQuiz(c *gin.Context) {
	userID, ok := ctxbase.GetUserID(c.Request.Context())
	if !ok {
		httpbase.Error(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	topicID := c.Param("topicId")

	var req quizSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpbase.Error(c, http.StatusBadRequest, "invalid input: "+err.Error())
		return
	}

	result, err := h.useCase.SubmitQuiz(c.Request.Context(), userID, topicID, appcore.QuizAnswerInput{Answers: req.Answers})
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to submit quiz")
		return
	}
	h.presenter.QuizResult(c, http.StatusOK, result)
}
```

- [x] **Step 5: Run handler tests to verify they pass**

```bash
cd server && go test ./internal/driving/httpui/handler/... -run TestVocab -v
```

Expected: `PASS` for all `TestVocabHandler_*` tests.

---

## Task 8: Route Registration + Module Wiring

**Files:**
- Modify: `server/internal/driving/httpui/server.go`
- Modify: `server/internal/driving/httpui/module.go`
- Modify: `server/cmd/api/main.go`

- [x] **Step 1: Update server.go — add vocab routes**

Replace the `RegisterRoutes` function signature and body in `server/internal/driving/httpui/server.go`:

```go
func RegisterRoutes(r *gin.Engine, authHandler *handler.Handler, vocabHandler *handler.VocabHandler, svc *domain.Service, cfg config.Config) {
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}

	// Public vocab endpoints
	r.GET("/api/topics", vocabHandler.ListTopics)
	r.GET("/api/topics/:id/words", vocabHandler.GetTopicWords)
	r.GET("/api/words/:id", vocabHandler.GetWord)

	// Protected vocab endpoints
	protected := r.Group("/api").Use(middleware.Auth(svc))
	{
		protected.GET("/review/due", vocabHandler.GetDueReviews)
		protected.POST("/review/:wordId", vocabHandler.SubmitReview)
		protected.GET("/quiz/:topicId", vocabHandler.GetQuiz)
		protected.POST("/quiz/:topicId", vocabHandler.SubmitQuiz)
	}
}
```

Also add the middleware import if not already present in the import block:

```go
import (
	// ... existing imports ...
	"github.com/sonpham/vielish/server/internal/driving/httpui/middleware"
)
```

- [x] **Step 2: Update module.go — add vocab handler/presenter providers**

Replace `server/internal/driving/httpui/module.go`:

```go
package httpui

import (
	"go.uber.org/fx"

	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
)

var Module = fx.Module("httpui",
	fx.Provide(NewGin),
	fx.Provide(handler.NewHandler),
	fx.Provide(handler.NewVocabHandler),
	fx.Provide(presenter.NewAuthPresenter),
	fx.Provide(presenter.NewVocabPresenter),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(RegisterLifecycle),
)
```

- [x] **Step 3: Update main.go — wire vocab modules**

Replace `server/cmd/api/main.go`:

```go
package main

import (
	"go.uber.org/fx"

	authAppcore "github.com/sonpham/vielish/server/internal/appcore/auth"
	vocabAppcore "github.com/sonpham/vielish/server/internal/appcore/vocab"
	authDomain "github.com/sonpham/vielish/server/internal/domain/auth"
	vocabDomain "github.com/sonpham/vielish/server/internal/domain/vocab"
	authDriven "github.com/sonpham/vielish/server/internal/driven/auth"
	"github.com/sonpham/vielish/server/internal/driven/database"
	vocabDriven "github.com/sonpham/vielish/server/internal/driven/vocab"
	"github.com/sonpham/vielish/server/internal/driving/httpui"
	"github.com/sonpham/vielish/server/pkg/config"
	pkglog "github.com/sonpham/vielish/server/pkg/log"
	"github.com/sonpham/vielish/server/pkg/tracing"
)

func main() {
	fx.New(
		// Infrastructure
		config.Module,
		pkglog.Module,
		tracing.Module,
		database.Module,

		// Auth feature
		fx.Module("auth",
			authDomain.Module,
			authAppcore.Module,
			authDriven.Module,
		),

		// Vocab feature
		fx.Module("vocab",
			vocabDomain.Module,
			vocabAppcore.Module,
			vocabDriven.Module,
		),

		// HTTP server
		httpui.Module,
	).Run()
}
```

- [x] **Step 4: Verify the binary compiles**

```bash
cd server && go build ./cmd/api/...
```

Expected: Exit 0, no errors.

- [x] **Step 5: Run all tests**

```bash
cd server && go test ./... -v
```

Expected: All tests pass (auth + vocab).

---

## Task 9: Frontend Vocab API Client

**Files:**
- Create: `web/lib/vocab-api.ts`

> **Prerequisite:** Read `node_modules/next/dist/docs/` for any Next.js 16 breaking changes that affect client-side code before implementing.

- [x] **Step 1: Implement vocab-api.ts**

`web/lib/vocab-api.ts`:
```typescript
import { api } from "./api";

export interface Topic {
  id: string;
  name: string;
  name_vi: string;
  description: string;
  level: string;
}

export interface Word {
  id: string;
  word: string;
  ipa_phonetic: string;
  part_of_speech: string;
  vi_meaning: string;
  en_definition: string;
  example_sentence: string;
  example_vi_translation: string;
  audio_url: string;
  image_url: string;
  level: string;
  topic_id: string;
}

export interface QuizQuestion {
  word_id: string;
  word: string;
  options: string[];
}

export interface QuizResult {
  score: number;
  total: number;
  results: { word_id: string; correct: boolean; correct_answer: string }[];
}

export async function fetchTopics(level?: string): Promise<Topic[]> {
  const query = level ? `?level=${level}` : "";
  const res = await api.request(`/api/topics${query}`);
  if (!res.ok) throw new Error("Failed to fetch topics");
  return res.json();
}

export async function fetchTopicWords(topicId: string): Promise<Word[]> {
  const res = await api.request(`/api/topics/${topicId}/words`);
  if (!res.ok) throw new Error("Failed to fetch words");
  return res.json();
}

export async function fetchWord(id: string): Promise<Word> {
  const res = await api.request(`/api/words/${id}`);
  if (!res.ok) throw new Error("Failed to fetch word");
  return res.json();
}

export async function fetchDueReviews(): Promise<Word[]> {
  const res = await api.request("/api/review/due");
  if (!res.ok) throw new Error("Failed to fetch due reviews");
  return res.json();
}

export async function submitReview(
  wordId: string,
  quality: 1 | 3 | 5
): Promise<void> {
  const res = await api.request(`/api/review/${wordId}`, {
    method: "POST",
    body: JSON.stringify({ quality }),
  });
  if (!res.ok) throw new Error("Failed to submit review");
}

export async function fetchQuiz(
  topicId: string
): Promise<{ questions: QuizQuestion[] }> {
  const res = await api.request(`/api/quiz/${topicId}`);
  if (!res.ok) throw new Error("Failed to fetch quiz");
  return res.json();
}

export async function submitQuiz(
  topicId: string,
  answers: { word_id: string; answer: string }[]
): Promise<QuizResult> {
  const res = await api.request(`/api/quiz/${topicId}`, {
    method: "POST",
    body: JSON.stringify({ answers }),
  });
  if (!res.ok) throw new Error("Failed to submit quiz");
  return res.json();
}
```

---

## Task 10: Frontend Topic List Page

**Files:**
- Create: `web/app/dashboard/topics/page.tsx`

> **Prerequisite:** Read `node_modules/next/dist/docs/` for any App Router breaking changes in Next.js 16.

- [x] **Step 1: Create topics list page**

`web/app/dashboard/topics/page.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { fetchTopics, Topic } from "@/lib/vocab-api";

const LEVELS = ["beginner", "intermediate", "advanced"];
const LEVEL_LABELS: Record<string, string> = {
  beginner: "Cơ bản",
  intermediate: "Trung cấp",
  advanced: "Nâng cao",
};

export default function TopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [level, setLevel] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchTopics(level || undefined)
      .then(setTopics)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [level]);

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Chủ đề từ vựng</h2>

      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setLevel("")}
          className={`px-3 py-1 rounded text-sm ${
            level === "" ? "bg-blue-600 text-white" : "bg-gray-200"
          }`}
        >
          Tất cả
        </button>
        {LEVELS.map((l) => (
          <button
            key={l}
            onClick={() => setLevel(l)}
            className={`px-3 py-1 rounded text-sm ${
              level === l ? "bg-blue-600 text-white" : "bg-gray-200"
            }`}
          >
            {LEVEL_LABELS[l]}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="text-gray-500">Đang tải...</p>
      ) : topics.length === 0 ? (
        <p className="text-gray-500">Chưa có chủ đề nào.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {topics.map((topic) => (
            <Link
              key={topic.id}
              href={`/dashboard/topics/${topic.id}`}
              className="block p-4 border rounded-lg hover:shadow-md transition-shadow"
            >
              <h3 className="font-semibold text-lg">{topic.name}</h3>
              <p className="text-gray-600 text-sm">{topic.name_vi}</p>
              {topic.description && (
                <p className="text-gray-500 text-sm mt-1">
                  {topic.description}
                </p>
              )}
              <span className="inline-block mt-2 px-2 py-0.5 bg-gray-100 text-xs rounded">
                {LEVEL_LABELS[topic.level] || topic.level}
              </span>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
```

---

## Task 11: Frontend Topic Detail + Flashcard

**Files:**
- Create: `web/app/dashboard/topics/[id]/page.tsx`
- Create: `web/components/flashcard.tsx`
- Create: `web/app/dashboard/topics/[id]/learn/page.tsx`

- [x] **Step 1: Create topic detail page**

`web/app/dashboard/topics/[id]/page.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { fetchTopicWords, Word } from "@/lib/vocab-api";

export default function TopicDetailPage() {
  const params = useParams<{ id: string }>();
  const [words, setWords] = useState<Word[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchTopicWords(params.id)
      .then(setWords)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [params.id]);

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <Link
          href="/dashboard/topics"
          className="text-blue-600 hover:underline text-sm"
        >
          ← Chủ đề
        </Link>
      </div>

      <div className="flex gap-3 mb-6">
        <Link
          href={`/dashboard/topics/${params.id}/learn`}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Học từ mới
        </Link>
        <Link
          href={`/dashboard/topics/${params.id}/quiz`}
          className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
        >
          Làm bài kiểm tra
        </Link>
      </div>

      <h3 className="text-lg font-semibold mb-3">
        Danh sách từ ({words.length})
      </h3>
      <div className="space-y-2">
        {words.map((word) => (
          <div key={word.id} className="p-3 border rounded flex justify-between">
            <div>
              <span className="font-medium">{word.word}</span>
              {word.ipa_phonetic && (
                <span className="text-gray-500 text-sm ml-2">
                  {word.ipa_phonetic}
                </span>
              )}
            </div>
            <span className="text-gray-600">{word.vi_meaning}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

- [x] **Step 2: Create flashcard component**

`web/components/flashcard.tsx`:
```tsx
"use client";

import { useState } from "react";
import { Word } from "@/lib/vocab-api";

interface FlashcardProps {
  word: Word;
  onRate: (quality: 1 | 3 | 5) => void;
}

export default function Flashcard({ word, onRate }: FlashcardProps) {
  const [flipped, setFlipped] = useState(false);

  return (
    <div className="max-w-md mx-auto">
      <div
        onClick={() => setFlipped(!flipped)}
        className="border rounded-lg p-8 min-h-[250px] flex flex-col items-center justify-center cursor-pointer hover:shadow-md transition-shadow"
      >
        {!flipped ? (
          <>
            <p className="text-3xl font-bold mb-2">{word.word}</p>
            {word.ipa_phonetic && (
              <p className="text-gray-500">{word.ipa_phonetic}</p>
            )}
            {word.part_of_speech && (
              <p className="text-gray-400 text-sm mt-1">
                {word.part_of_speech}
              </p>
            )}
            <p className="text-gray-400 text-sm mt-4">Nhấn để lật thẻ</p>
          </>
        ) : (
          <>
            <p className="text-2xl font-semibold text-blue-700 mb-3">
              {word.vi_meaning}
            </p>
            {word.en_definition && (
              <p className="text-gray-600 text-sm mb-2">
                {word.en_definition}
              </p>
            )}
            {word.example_sentence && (
              <div className="mt-3 text-sm text-gray-500">
                <p className="italic">{word.example_sentence}</p>
                {word.example_vi_translation && (
                  <p className="text-gray-400 mt-1">
                    {word.example_vi_translation}
                  </p>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {flipped && (
        <div className="flex justify-center gap-3 mt-4">
          <button
            onClick={() => onRate(1)}
            className="px-6 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Khó
          </button>
          <button
            onClick={() => onRate(3)}
            className="px-6 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600"
          >
            Ổn
          </button>
          <button
            onClick={() => onRate(5)}
            className="px-6 py-2 bg-green-500 text-white rounded hover:bg-green-600"
          >
            Dễ
          </button>
        </div>
      )}
    </div>
  );
}
```

- [x] **Step 3: Create learn page**

`web/app/dashboard/topics/[id]/learn/page.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import Flashcard from "@/components/flashcard";
import { fetchTopicWords, submitReview, Word } from "@/lib/vocab-api";

export default function LearnPage() {
  const params = useParams<{ id: string }>();
  const [words, setWords] = useState<Word[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [finished, setFinished] = useState(false);

  useEffect(() => {
    fetchTopicWords(params.id)
      .then(setWords)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [params.id]);

  const handleRate = async (quality: 1 | 3 | 5) => {
    const word = words[currentIndex];
    try {
      await submitReview(word.id, quality);
    } catch (err) {
      console.error("Review submit failed:", err);
    }

    if (currentIndex + 1 < words.length) {
      setCurrentIndex(currentIndex + 1);
    } else {
      setFinished(true);
    }
  };

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  if (words.length === 0) {
    return <p className="text-gray-500">Chưa có từ nào trong chủ đề này.</p>;
  }

  if (finished) {
    return (
      <div className="text-center py-12">
        <p className="text-2xl font-bold mb-4">Hoàn thành! 🎉</p>
        <p className="text-gray-600 mb-6">
          Bạn đã học xong {words.length} từ.
        </p>
        <div className="flex justify-center gap-3">
          <Link
            href={`/dashboard/topics/${params.id}`}
            className="px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
          >
            Quay lại
          </Link>
          <Link
            href={`/dashboard/topics/${params.id}/quiz`}
            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            Làm bài kiểm tra
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <Link
          href={`/dashboard/topics/${params.id}`}
          className="text-blue-600 hover:underline text-sm"
        >
          ← Quay lại
        </Link>
        <span className="text-sm text-gray-500">
          {currentIndex + 1} / {words.length}
        </span>
      </div>

      <Flashcard word={words[currentIndex]} onRate={handleRate} />
    </div>
  );
}
```

---

## Task 12: Frontend Quiz Page

**Files:**
- Create: `web/components/quiz-question.tsx`
- Create: `web/app/dashboard/topics/[id]/quiz/page.tsx`

- [x] **Step 1: Create quiz question component**

`web/components/quiz-question.tsx`:
```tsx
"use client";

import { useState } from "react";
import { QuizQuestion } from "@/lib/vocab-api";

interface QuizQuestionCardProps {
  question: QuizQuestion;
  index: number;
  onAnswer: (wordId: string, answer: string) => void;
  selectedAnswer?: string;
}

export default function QuizQuestionCard({
  question,
  index,
  onAnswer,
  selectedAnswer,
}: QuizQuestionCardProps) {
  return (
    <div className="border rounded-lg p-4 mb-4">
      <p className="font-medium mb-3">
        {index + 1}. <span className="text-lg">{question.word}</span> nghĩa là
        gì?
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {question.options.map((option) => (
          <button
            key={option}
            onClick={() => onAnswer(question.word_id, option)}
            className={`p-2 text-left rounded border ${
              selectedAnswer === option
                ? "bg-blue-100 border-blue-500"
                : "hover:bg-gray-50"
            }`}
          >
            {option}
          </button>
        ))}
      </div>
    </div>
  );
}
```

- [x] **Step 2: Create quiz page**

`web/app/dashboard/topics/[id]/quiz/page.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import QuizQuestionCard from "@/components/quiz-question";
import {
  fetchQuiz,
  submitQuiz,
  QuizQuestion,
  QuizResult,
} from "@/lib/vocab-api";

export default function QuizPage() {
  const params = useParams<{ id: string }>();
  const [questions, setQuestions] = useState<QuizQuestion[]>([]);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [result, setResult] = useState<QuizResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchQuiz(params.id)
      .then((data) => setQuestions(data.questions))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [params.id]);

  const handleAnswer = (wordId: string, answer: string) => {
    if (result) return;
    setAnswers((prev) => ({ ...prev, [wordId]: answer }));
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    try {
      const answerList = Object.entries(answers).map(([word_id, answer]) => ({
        word_id,
        answer,
      }));
      const res = await submitQuiz(params.id, answerList);
      setResult(res);
    } catch (err) {
      console.error("Quiz submit failed:", err);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <Link
          href={`/dashboard/topics/${params.id}`}
          className="text-blue-600 hover:underline text-sm"
        >
          ← Quay lại
        </Link>
        <h2 className="text-xl font-bold">Bài kiểm tra</h2>
      </div>

      {questions.map((q, i) => (
        <QuizQuestionCard
          key={q.word_id}
          question={q}
          index={i}
          onAnswer={handleAnswer}
          selectedAnswer={answers[q.word_id]}
        />
      ))}

      {!result ? (
        <button
          onClick={handleSubmit}
          disabled={
            submitting || Object.keys(answers).length !== questions.length
          }
          className="mt-4 px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {submitting ? "Đang nộp..." : "Nộp bài"}
        </button>
      ) : (
        <div className="mt-6 p-4 border rounded-lg bg-gray-50">
          <p className="text-xl font-bold mb-2">
            Kết quả: {result.score}/{result.total}
          </p>
          <div className="space-y-1">
            {result.results.map((r) => (
              <p
                key={r.word_id}
                className={r.correct ? "text-green-600" : "text-red-600"}
              >
                {r.correct ? "✓" : "✗"} Đáp án đúng: {r.correct_answer}
              </p>
            ))}
          </div>
          <Link
            href={`/dashboard/topics/${params.id}`}
            className="inline-block mt-4 px-4 py-2 bg-gray-200 rounded hover:bg-gray-300"
          >
            Quay lại chủ đề
          </Link>
        </div>
      )}
    </div>
  );
}
```

---

## Task 13: Frontend SRS Review Page

**Files:**
- Create: `web/app/dashboard/review/page.tsx`

- [x] **Step 1: Create review page**

`web/app/dashboard/review/page.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Flashcard from "@/components/flashcard";
import { fetchDueReviews, submitReview, Word } from "@/lib/vocab-api";

export default function ReviewPage() {
  const [words, setWords] = useState<Word[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [finished, setFinished] = useState(false);

  useEffect(() => {
    fetchDueReviews()
      .then(setWords)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const handleRate = async (quality: 1 | 3 | 5) => {
    const word = words[currentIndex];
    try {
      await submitReview(word.id, quality);
    } catch (err) {
      console.error("Review submit failed:", err);
    }

    if (currentIndex + 1 < words.length) {
      setCurrentIndex(currentIndex + 1);
    } else {
      setFinished(true);
    }
  };

  if (loading) {
    return <p className="text-gray-500">Đang tải...</p>;
  }

  if (words.length === 0 && !finished) {
    return (
      <div className="text-center py-12">
        <p className="text-xl font-semibold mb-2">Không có từ cần ôn tập</p>
        <p className="text-gray-500 mb-4">
          Hãy học thêm từ mới hoặc quay lại sau.
        </p>
        <Link
          href="/dashboard/topics"
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Xem chủ đề
        </Link>
      </div>
    );
  }

  if (finished) {
    return (
      <div className="text-center py-12">
        <p className="text-2xl font-bold mb-4">Ôn tập xong!</p>
        <p className="text-gray-600 mb-6">
          Bạn đã ôn tập {words.length} từ hôm nay.
        </p>
        <Link
          href="/dashboard"
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Về trang chủ
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold">Ôn tập từ vựng</h2>
        <span className="text-sm text-gray-500">
          {currentIndex + 1} / {words.length}
        </span>
      </div>

      <Flashcard word={words[currentIndex]} onRate={handleRate} />
    </div>
  );
}
```

---

## Task 14: Dashboard Integration

**Files:**
- Modify: `web/app/dashboard/layout.tsx`
- Modify: `web/app/dashboard/page.tsx`

- [x] **Step 1: Update dashboard layout — add nav links**

Add topic and review links to the `<nav>` in `web/app/dashboard/layout.tsx`. Replace the nav bar section:

```tsx
<nav className="border-b px-6 py-4 flex justify-between items-center">
  <div className="flex items-center gap-6">
    <h1 className="text-xl font-bold">Vielish</h1>
    <Link href="/dashboard" className="text-sm text-gray-600 hover:text-gray-900">
      Trang chủ
    </Link>
    <Link href="/dashboard/topics" className="text-sm text-gray-600 hover:text-gray-900">
      Chủ đề
    </Link>
    <Link href="/dashboard/review" className="text-sm text-gray-600 hover:text-gray-900">
      Ôn tập
    </Link>
  </div>
  <button
    onClick={() => {
      logout();
      router.push("/login");
    }}
    className="text-sm text-gray-600 hover:text-gray-900"
  >
    Đăng xuất
  </button>
</nav>
```

Add `Link` import at the top:

```tsx
import Link from "next/link";
```

- [x] **Step 2: Update dashboard page — add vocab links**

Replace `web/app/dashboard/page.tsx`:

```tsx
import Link from "next/link";

export default function DashboardPage() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Bảng điều khiển</h2>
      <p className="text-gray-600 mb-6">
        Chào mừng bạn đến với Vielish! Chọn một hoạt động để bắt đầu.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Link
          href="/dashboard/topics"
          className="block p-6 border rounded-lg hover:shadow-md transition-shadow"
        >
          <h3 className="text-lg font-semibold mb-1">Chủ đề từ vựng</h3>
          <p className="text-gray-500 text-sm">
            Học từ mới theo chủ đề với flashcard và SRS.
          </p>
        </Link>
        <Link
          href="/dashboard/review"
          className="block p-6 border rounded-lg hover:shadow-md transition-shadow"
        >
          <h3 className="text-lg font-semibold mb-1">Ôn tập hôm nay</h3>
          <p className="text-gray-500 text-sm">
            Ôn lại các từ đã học theo lịch SRS.
          </p>
        </Link>
      </div>
    </div>
  );
}
```

---

## Task 15: Smoke Test

- [x] **Step 1: Start infrastructure**

```bash
docker-compose up -d postgres redis
```

Expected: Both containers running.

- [x] **Step 2: Apply migrations**

```bash
psql "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" \
  -f server/migrations/002_create_topics.up.sql \
  -f server/migrations/003_create_words.up.sql \
  -f server/migrations/004_create_user_word_progress.up.sql
```

Expected: Tables created.

- [x] **Step 3: Seed sample data**

```bash
psql "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" -c "
INSERT INTO topics (name, name_vi, description, level) VALUES
  ('Food', 'Thức ăn', 'Từ vựng về thức ăn và đồ uống', 'beginner'),
  ('Travel', 'Du lịch', 'Từ vựng về du lịch và phương tiện', 'intermediate');

INSERT INTO words (word, ipa_phonetic, part_of_speech, vi_meaning, en_definition, example_sentence, example_vi_translation, level, topic_id)
SELECT 'apple', '/ˈæp.əl/', 'noun', 'quả táo', 'A round fruit with red or green skin', 'I eat an apple every day.', 'Tôi ăn một quả táo mỗi ngày.', 'beginner', id FROM topics WHERE name = 'Food'
UNION ALL
SELECT 'rice', '/raɪs/', 'noun', 'cơm / gạo', 'A cereal grain that is a staple food', 'Rice is the main food in Vietnam.', 'Cơm là thức ăn chính ở Việt Nam.', 'beginner', id FROM topics WHERE name = 'Food'
UNION ALL
SELECT 'water', '/ˈwɔː.tər/', 'noun', 'nước', 'A clear liquid essential for life', 'Please give me a glass of water.', 'Làm ơn cho tôi một cốc nước.', 'beginner', id FROM topics WHERE name = 'Food'
UNION ALL
SELECT 'bread', '/bred/', 'noun', 'bánh mì', 'Food made from flour, water, and yeast', 'I bought some bread from the bakery.', 'Tôi mua bánh mì từ tiệm bánh.', 'beginner', id FROM topics WHERE name = 'Food';
"
```

Expected: 2 topics, 4 words inserted.

- [x] **Step 4: Run the server**

```bash
cd server && go run cmd/api/main.go
```

Expected: Server starts on port 8080.

- [x] **Step 5: List topics**

```bash
curl -s http://localhost:8080/api/topics | jq .
```

Expected: JSON array with 2 topics.

- [x] **Step 6: List words in topic**

```bash
TOPIC_ID=$(curl -s http://localhost:8080/api/topics | jq -r '.[0].id')
curl -s "http://localhost:8080/api/topics/$TOPIC_ID/words" | jq .
```

Expected: JSON array with 4 words.

- [x] **Step 7: Submit review (authenticated)**

```bash
# Register and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"testpass123","display_name":"Test"}' | jq -r '.access_token')

# Get a word ID
WORD_ID=$(curl -s "http://localhost:8080/api/topics/$TOPIC_ID/words" | jq -r '.[0].id')

# Submit review
curl -s -X POST "http://localhost:8080/api/review/$WORD_ID" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"quality":5}' | jq .
```

Expected: `{"message": "review submitted"}`

- [x] **Step 8: Get quiz**

```bash
curl -s "http://localhost:8080/api/quiz/$TOPIC_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Expected: JSON with `questions` array, each having 4 options.

- [x] **Step 9: Get due reviews**

```bash
curl -s http://localhost:8080/api/review/due \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Expected: JSON array (may be empty if next_review_at is in the future, or contains the reviewed word if due).

- [x] **Step 10: Frontend verification**

```bash
cd web && npm run dev
```

Open `http://localhost:3000/dashboard/topics` — should show topic list.
Navigate to a topic — should show word list with Learn/Quiz buttons.
Click "Học từ mới" — flashcard view with flip and rating.
Click "Làm bài kiểm tra" — quiz with multiple choice questions.
Visit `/dashboard/review` — SRS review page.
