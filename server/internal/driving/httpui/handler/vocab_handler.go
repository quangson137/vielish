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
	GetStats(ctx context.Context, userID string) (*appcore.StatsOutput, error)
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

func (h *VocabHandler) GetStats(c *gin.Context) {
	userID, ok := ctxbase.GetUserID(c.Request.Context())
	if !ok {
		httpbase.Error(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	stats, err := h.useCase.GetStats(c.Request.Context(), userID)
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to get stats")
		return
	}
	h.presenter.Stats(c, http.StatusOK, stats)
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
