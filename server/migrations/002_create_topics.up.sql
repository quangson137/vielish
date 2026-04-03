CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    name_vi VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    level user_level NOT NULL DEFAULT 'beginner'
);

CREATE INDEX idx_topics_level ON topics(level);
