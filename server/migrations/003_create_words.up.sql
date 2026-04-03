CREATE TABLE words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(200) NOT NULL,
    ipa_phonetic VARCHAR(200) NOT NULL DEFAULT '',
    part_of_speech VARCHAR(50) NOT NULL DEFAULT '',
    vi_meaning VARCHAR(500) NOT NULL,
    en_definition TEXT NOT NULL DEFAULT '',
    example_sentence TEXT NOT NULL DEFAULT '',
    example_vi_translation TEXT NOT NULL DEFAULT '',
    audio_url TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    level user_level NOT NULL DEFAULT 'beginner',
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE
);

CREATE INDEX idx_words_topic_id ON words(topic_id);
CREATE INDEX idx_words_level ON words(level);
