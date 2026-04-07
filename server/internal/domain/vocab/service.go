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
	ef += 0.1 - (5-q)*(0.08+(5-q)*0.02)
	if ef < 1.3 {
		ef = 1.3
	}

	return ReviewResult{
		EaseFactor:   ef,
		IntervalDays: interval,
		NextReviewAt: now.Add(time.Duration(interval) * 24 * time.Hour),
	}
}
