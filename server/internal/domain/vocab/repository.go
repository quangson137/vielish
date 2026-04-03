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

	// Stats
	CountLearnedWords(ctx context.Context, userID string) (int, error)
	CountDueWords(ctx context.Context, userID string, now time.Time) (int, error)
	GetReviewDates(ctx context.Context, userID string) ([]time.Time, error)
}
