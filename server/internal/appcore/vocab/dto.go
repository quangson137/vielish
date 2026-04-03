package appcore

type TopicOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameVI      string `json:"name_vi"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

type WordOutput struct {
	ID                   string `json:"id"`
	Word                 string `json:"word"`
	IPAPhonetic          string `json:"ipa_phonetic"`
	PartOfSpeech         string `json:"part_of_speech"`
	VIMeaning            string `json:"vi_meaning"`
	ENDefinition         string `json:"en_definition"`
	ExampleSentence      string `json:"example_sentence"`
	ExampleVITranslation string `json:"example_vi_translation"`
	AudioURL             string `json:"audio_url"`
	ImageURL             string `json:"image_url"`
	Level                string `json:"level"`
	TopicID              string `json:"topic_id"`
}

type ReviewInput struct {
	Quality int `json:"quality"` // 1 (Hard), 3 (OK), 5 (Easy)
}

type QuizQuestion struct {
	WordID  string   `json:"word_id"`
	Word    string   `json:"word"`
	Options []string `json:"options"`
}

type QuizAnswerInput struct {
	Answers []QuizAnswer `json:"answers"`
}

type QuizAnswer struct {
	WordID string `json:"word_id"`
	Answer string `json:"answer"`
}

type QuizResult struct {
	Score   int              `json:"score"`
	Total   int              `json:"total"`
	Results []QuizItemResult `json:"results"`
}

type QuizItemResult struct {
	WordID        string `json:"word_id"`
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correct_answer"`
}
