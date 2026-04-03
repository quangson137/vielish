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
