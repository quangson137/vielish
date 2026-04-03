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
