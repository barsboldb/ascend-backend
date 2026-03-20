CREATE TABLE program_exercises (
  id               UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
  program_day_id   UUID  NOT NULL,
  exercise_id      UUID  NOT NULL,    
  position         INT   NOT NULL,
  sets             INT   NOT NULL,
  rep_min          INT   NOT NULL,
  rep_max          INT   NOT NULL,
  weight_increment FLOAT,
  FOREIGN KEY (program_day_id) REFERENCES program_days(id),
  FOREIGN KEY (exercise_id) REFERENCES exercises(id)
);
