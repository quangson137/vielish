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
	for _, p := range r.progress {
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

func (r *stubRepo) CountLearnedWords(_ context.Context, userID string) (int, error) {
	count := 0
	for _, p := range r.progress {
		if p.UserID == userID && p.ReviewCount > 0 {
			count++
		}
	}
	return count, nil
}

func (r *stubRepo) CountDueWords(_ context.Context, userID string, now time.Time) (int, error) {
	count := 0
	for _, p := range r.progress {
		if p.UserID == userID && !p.NextReviewAt.After(now) {
			count++
		}
	}
	return count, nil
}

func (r *stubRepo) GetReviewDates(_ context.Context, userID string) ([]time.Time, error) {
	return nil, nil
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

func TestGetWord(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{{ID: "w1", Word: "apple", VIMeaning: "quả táo", TopicID: "t1"}}
	uc := appcore.NewUseCase(repo, domain.NewService())

	out, err := uc.GetWord(context.Background(), "w1")
	if err != nil {
		t.Fatalf("GetWord error: %v", err)
	}
	if out == nil {
		t.Fatal("expected word output, got nil")
	}
	if out.Word != "apple" {
		t.Errorf("Word = %q, want %q", out.Word, "apple")
	}
}

func TestGetWord_NotFound(t *testing.T) {
	repo := newStubRepo()
	uc := appcore.NewUseCase(repo, domain.NewService())

	_, err := uc.GetWord(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing word")
	}
}

func TestGetDueReviews(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{{ID: "w1", Word: "apple", VIMeaning: "quả táo", TopicID: "t1"}}
	past := time.Now().Add(-24 * time.Hour)
	repo.progress["user-1:w1"] = &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "w1",
		NextReviewAt: past,
		ReviewCount:  1,
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	out, err := uc.GetDueReviews(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetDueReviews error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].ID != "w1" {
		t.Errorf("ID = %q, want %q", out[0].ID, "w1")
	}
}

func TestGetDueReviews_NoItems(t *testing.T) {
	repo := newStubRepo()
	uc := appcore.NewUseCase(repo, domain.NewService())

	out, err := uc.GetDueReviews(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetDueReviews error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("len = %d, want 0", len(out))
	}
}

func TestSubmitReview_UpdatesExistingProgress(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{{ID: "w1", Word: "apple", TopicID: "t1"}}
	now := time.Now()
	repo.progress["user-1:w1"] = &domain.UserWordProgress{
		UserID:       "user-1",
		WordID:       "w1",
		EaseFactor:   2.5,
		IntervalDays: 1,
		ReviewCount:  1,
		NextReviewAt: now,
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	err := uc.SubmitReview(context.Background(), "user-1", "w1", appcore.ReviewInput{Quality: 3})
	if err != nil {
		t.Fatalf("SubmitReview error: %v", err)
	}

	p, _ := repo.GetProgress(context.Background(), "user-1", "w1")
	if p.ReviewCount != 2 {
		t.Errorf("ReviewCount = %d, want 2", p.ReviewCount)
	}
	// Second review with quality=3 from interval=1 → interval=6
	if p.IntervalDays != 6 {
		t.Errorf("IntervalDays = %d, want 6", p.IntervalDays)
	}
}

func TestGetStats(t *testing.T) {
	repo := newStubRepo()
	repo.words = []domain.Word{
		{ID: "w1", TopicID: "t1"},
		{ID: "w2", TopicID: "t1"},
	}
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(24 * time.Hour)
	repo.progress["user-1:w1"] = &domain.UserWordProgress{
		UserID: "user-1", WordID: "w1", ReviewCount: 1, NextReviewAt: past,
	}
	repo.progress["user-1:w2"] = &domain.UserWordProgress{
		UserID: "user-1", WordID: "w2", ReviewCount: 2, NextReviewAt: future,
	}
	uc := appcore.NewUseCase(repo, domain.NewService())

	stats, err := uc.GetStats(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if stats.TotalLearned != 2 {
		t.Errorf("TotalLearned = %d, want 2", stats.TotalLearned)
	}
	if stats.DueToday != 1 {
		t.Errorf("DueToday = %d, want 1", stats.DueToday)
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
