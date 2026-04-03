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
func (s *stubVocabUC) GetStats(_ context.Context, _ string) (*appcore.StatsOutput, error) {
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
