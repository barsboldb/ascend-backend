CREATE TABLE programs (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT,
  total_weeks INT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
