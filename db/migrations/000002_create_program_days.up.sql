CREATE TABLE program_days (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  program_id  UUID NOT NULL,
  week_number INT NOT NULL,
  day_number  INT NOT NULL,
  label       TEXT NOT NULL,
  FOREIGN KEY (program_id) REFERENCES programs(id)
);
