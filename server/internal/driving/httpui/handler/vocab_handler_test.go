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
	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
	"github.com/sonpham/vielish/server/internal/driving/httpui/handler"
	"github.com/sonpham/vielish/server/internal/driving/httpui/presenter"
	"github.com/sonpham/vielish/server/pkg/ctxbase"
)

// --- Stub ---

type stubVocabUC struct {
	topics        []appcore.TopicOutput
	words         []appcore.WordOutput
	word          *appcore.WordOutput
	questions     []appcore.QuizQuestion
	quiz          *appcore.QuizResult
	stats         *appcore.StatsOutput
	reviewErr     error
	topicWordsErr error
	wordErr       error
}

func (s *stubVocabUC) ListTopics(_ context.Context, _ string) ([]appcore.TopicOutput, error) {
	return s.topics, nil
}
func (s *stubVocabUC) GetTopicWords(_ context.Context, _ string) ([]appcore.WordOutput, error) {
	return s.words, s.topicWordsErr
}
func (s *stubVocabUC) GetWord(_ context.Context, _ string) (*appcore.WordOutput, error) {
	if s.wordErr != nil {
		return nil, s.wordErr
	}
	if s.word != nil {
		return s.word, nil
	}
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
func (s *stubVocabUC) GetStats(_ context.Context, _ string) (*appcore.StatsOutput, error) {
	if s.stats != nil {
		return s.stats, nil
	}
	return &appcore.StatsOutput{}, nil
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

func TestVocabHandler_GetTopicWords(t *testing.T) {
	stub := &stubVocabUC{
		words: []appcore.WordOutput{
			{ID: "w1", Word: "apple", VIMeaning: "quả táo"},
			{ID: "w2", Word: "rice", VIMeaning: "cơm"},
		},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/topics/t1/words", nil)
	c.Params = gin.Params{{Key: "id", Value: "t1"}}

	h.GetTopicWords(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body []map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if len(body) != 2 {
		t.Fatalf("len = %d, want 2", len(body))
	}
}

func TestVocabHandler_GetTopicWords_404(t *testing.T) {
	stub := &stubVocabUC{
		topicWordsErr: domain.ErrTopicNotFound,
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/topics/unknown/words", nil)
	c.Params = gin.Params{{Key: "id", Value: "unknown"}}

	h.GetTopicWords(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestVocabHandler_GetWord(t *testing.T) {
	stub := &stubVocabUC{
		word: &appcore.WordOutput{ID: "w1", Word: "apple", VIMeaning: "quả táo"},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/words/w1", nil)
	c.Params = gin.Params{{Key: "id", Value: "w1"}}

	h.GetWord(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["word"] != "apple" {
		t.Errorf("word = %v, want apple", body["word"])
	}
}

func TestVocabHandler_GetWord_404(t *testing.T) {
	stub := &stubVocabUC{
		wordErr: domain.ErrWordNotFound,
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/words/unknown", nil)
	c.Params = gin.Params{{Key: "id", Value: "unknown"}}

	h.GetWord(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestVocabHandler_GetDueReviews(t *testing.T) {
	stub := &stubVocabUC{
		words: []appcore.WordOutput{{ID: "w1", Word: "apple"}},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/review/due", nil)
	setUserID(c, "user-1")

	h.GetDueReviews(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestVocabHandler_GetDueReviews_401_NoUser(t *testing.T) {
	stub := &stubVocabUC{}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/review/due", nil)
	// no user injected into context

	h.GetDueReviews(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestVocabHandler_SubmitReview_400_InvalidInput(t *testing.T) {
	stub := &stubVocabUC{}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/review/w1", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "wordId", Value: "w1"}}
	setUserID(c, "user-1")

	h.SubmitReview(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestVocabHandler_SubmitReview_404_WordNotFound(t *testing.T) {
	stub := &stubVocabUC{
		reviewErr: domain.ErrWordNotFound,
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/review/unknown", bytes.NewBufferString(`{"quality":5}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "wordId", Value: "unknown"}}
	setUserID(c, "user-1")

	h.SubmitReview(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestVocabHandler_SubmitQuiz(t *testing.T) {
	stub := &stubVocabUC{
		quiz: &appcore.QuizResult{Score: 2, Total: 2},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	body := `{"answers":[{"word_id":"w1","answer":"quả táo"}]}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/quiz/t1", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "topicId", Value: "t1"}}
	setUserID(c, "user-1")

	h.SubmitQuiz(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var respBody map[string]any
	json.NewDecoder(w.Body).Decode(&respBody)
	if respBody["score"] != float64(2) {
		t.Errorf("score = %v, want 2", respBody["score"])
	}
}

func TestVocabHandler_GetStats(t *testing.T) {
	stub := &stubVocabUC{
		stats: &appcore.StatsOutput{Streak: 3, TotalLearned: 10, DueToday: 5},
	}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats", nil)
	setUserID(c, "user-1")

	h.GetStats(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["streak"] != float64(3) {
		t.Errorf("streak = %v, want 3", body["streak"])
	}
}

func TestVocabHandler_GetStats_401_NoUser(t *testing.T) {
	stub := &stubVocabUC{}
	h := handler.NewVocabHandlerFromInterface(stub, presenter.NewVocabPresenter())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats", nil)
	// no user in context

	h.GetStats(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
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
	var respBody map[string]any
	json.NewDecoder(w.Body).Decode(&respBody)
	questions, ok := respBody["questions"].([]any)
	if !ok || len(questions) != 1 {
		t.Fatalf("questions count = %v, want 1", respBody["questions"])
	}
}
