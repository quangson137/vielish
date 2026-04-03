package driven

import (
	"time"

	domain "github.com/sonpham/vielish/server/internal/domain/vocab"
)

type TopicModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string `gorm:"not null"`
	NameVI      string `gorm:"column:name_vi;not null"`
	Description string
	Level       string `gorm:"not null"`
}

func (TopicModel) TableName() string { return "topics" }

func (m *TopicModel) ToEntity() *domain.Topic {
	return &domain.Topic{
		ID:          m.ID,
		Name:        m.Name,
		NameVI:      m.NameVI,
		Description: m.Description,
		Level:       m.Level,
	}
}

type WordModel struct {
	ID                   string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Word                 string `gorm:"not null"`
	IPAPhonetic          string `gorm:"column:ipa_phonetic"`
	PartOfSpeech         string `gorm:"column:part_of_speech"`
	VIMeaning            string `gorm:"column:vi_meaning;not null"`
	ENDefinition         string `gorm:"column:en_definition"`
	ExampleSentence      string `gorm:"column:example_sentence"`
	ExampleVITranslation string `gorm:"column:example_vi_translation"`
	AudioURL             string `gorm:"column:audio_url"`
	ImageURL             string `gorm:"column:image_url"`
	Level                string `gorm:"not null"`
	TopicID              string `gorm:"column:topic_id;type:uuid;not null"`
}

func (WordModel) TableName() string { return "words" }

func (m *WordModel) ToEntity() *domain.Word {
	return &domain.Word{
		ID:                   m.ID,
		Word:                 m.Word,
		IPAPhonetic:          m.IPAPhonetic,
		PartOfSpeech:         m.PartOfSpeech,
		VIMeaning:            m.VIMeaning,
		ENDefinition:         m.ENDefinition,
		ExampleSentence:      m.ExampleSentence,
		ExampleVITranslation: m.ExampleVITranslation,
		AudioURL:             m.AudioURL,
		ImageURL:             m.ImageURL,
		Level:                m.Level,
		TopicID:              m.TopicID,
	}
}

type UserWordProgressModel struct {
	UserID         string     `gorm:"primaryKey;type:uuid"`
	WordID         string     `gorm:"primaryKey;type:uuid"`
	EaseFactor     float64    `gorm:"column:ease_factor;default:2.5"`
	IntervalDays   int        `gorm:"column:interval_days;default:0"`
	NextReviewAt   time.Time  `gorm:"column:next_review_at"`
	ReviewCount    int        `gorm:"column:review_count;default:0"`
	LastReviewedAt *time.Time `gorm:"column:last_reviewed_at"`
}

func (UserWordProgressModel) TableName() string { return "user_word_progress" }

func (m *UserWordProgressModel) ToEntity() *domain.UserWordProgress {
	return &domain.UserWordProgress{
		UserID:         m.UserID,
		WordID:         m.WordID,
		EaseFactor:     m.EaseFactor,
		IntervalDays:   m.IntervalDays,
		NextReviewAt:   m.NextReviewAt,
		ReviewCount:    m.ReviewCount,
		LastReviewedAt: m.LastReviewedAt,
	}
}
