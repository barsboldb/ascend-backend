CREATE TABLE exercise_sets (
  id          UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id  UUID       NOT NULL,
  exercise_id UUID       NOT NULL,
  set_number  INT        NOT NULL,
  weight_kg   FLOAT      NOT NULL,
  reps        INT        NOT NULL,
  failure     BOOLEAN    DEFAULT FALSE,
  logged_at   TIMESTAMPTZ DEFAULT NOW(),
  FOREIGN KEY (session_id) REFERENCES sessions(id),
  FOREIGN KEY (exercise_id) REFERENCES exercises(id)
);
