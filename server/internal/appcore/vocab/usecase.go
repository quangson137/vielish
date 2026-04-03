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
