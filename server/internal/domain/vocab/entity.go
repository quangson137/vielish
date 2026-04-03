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
