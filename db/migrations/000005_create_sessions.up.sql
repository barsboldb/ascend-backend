CREATE TABLE sessions (
  id             UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
  program_day_id UUID       NOT NULL,
  week_number    INT        NOT NULL,
  started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ended_at       TIMESTAMPTZ,
  notes          TEXT,
  FOREIGN KEY (program_day_id) REFERENCES program_days(id)
);
