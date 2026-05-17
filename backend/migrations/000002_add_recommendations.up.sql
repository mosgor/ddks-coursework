CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE events ADD COLUMN embedding vector(384);

CREATE TABLE IF NOT EXISTS user_behavior (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    event_id INT REFERENCES events(id) ON DELETE CASCADE,
    event_type VARCHAR(20) NOT NULL, -- 'view', 'cart_add', 'purchase'
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_behavior_user_type ON user_behavior(user_id, event_type);
CREATE INDEX IF NOT EXISTS idx_events_embedding ON events USING ivfflat (embedding vector_cosine_ops) WITH (lists = 10);

ALTER TABLE events 
  ADD COLUMN IF NOT EXISTS genre VARCHAR(50),
  ADD COLUMN IF NOT EXISTS tags JSONB;