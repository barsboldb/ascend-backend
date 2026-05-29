CREATE TABLE users (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(255),
  created_at    TIMESTAMPTZ  DEFAULT NOW()
);

ALTER TABLE sessions ADD COLUMN user_id UUID REFERENCES users(id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
