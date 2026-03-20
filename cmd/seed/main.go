package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exerciseDef struct {
	name, muscleGroup, equipment string
}

// sg represents a group of identical sets
type sg struct {
	exercise string
	weightKg float64
	reps     int
	numSets  int
	failure  bool // if true, marks the last set of this group as failure
}

type sessionDef struct {
	date         time.Time
	dayKey       string // "push_a", "pull_a", "leg_a", "push_b", "pull_b", "leg_b"
	weekNum      int
	durationMins int
	sets         []sg
}

var exerciseDefs = []exerciseDef{
	// Legs
	{"Back Squat", "Legs", "Barbell"},
	{"Front Squat", "Legs", "Barbell"},
	{"Deadlift", "Back", "Barbell"},
	{"Conventional Deadlift", "Legs", "Barbell"},
	{"Romanian Deadlift", "Back", "Barbell"},
	{"Leg Press", "Legs", "Machine"},
	{"Leg Curls", "Legs", "Machine"},
	{"Walking Lunges", "Legs", "Dumbbell"},
	{"Standing Calf Raises", "Legs", "Machine"},
	{"Seated Calf Raises", "Legs", "Machine"},
	// Core
	{"Hanging Leg Raises", "Core", "Bodyweight"},
	{"Dead Bug", "Core", "Bodyweight"},
	{"Plank", "Core", "Bodyweight"},
	{"Pallof Press", "Core", "Cable"},
	// Chest
	{"Bench Press", "Chest", "Barbell"},
	{"Incline Dumbbell Press", "Chest", "Dumbbell"},
	{"Flat Dumbbell Press", "Chest", "Dumbbell"},
	{"Cable Flyes", "Chest", "Cable"},
	// Shoulders
	{"Overhead Barbell Press", "Shoulders", "Barbell"},
	{"Shoulder Press", "Shoulders", "Dumbbell"},
	{"Lateral Raises", "Shoulders", "Dumbbell"},
	{"Face Pulls", "Shoulders", "Cable"},
	{"Rear Delt Fly", "Shoulders", "Dumbbell"},
	// Triceps
	{"Skull Crushers", "Triceps", "Barbell"},
	{"Close Grip Bench Press", "Triceps", "Barbell"},
	{"Cable Pushdown", "Triceps", "Cable"},
	{"Overhead Tricep Extension", "Triceps", "Dumbbell"},
	{"Close Grip Push Ups", "Triceps", "Bodyweight"},
	{"Diamond Push Ups", "Triceps", "Bodyweight"},
	// Back
	{"Barbell Row", "Back", "Barbell"},
	{"Lat Pulldown", "Back", "Machine"},
	{"Pull Up", "Back", "Bodyweight"},
	{"Assisted Pull Up", "Back", "Machine"},
	{"Seated Cable Row", "Back", "Cable"},
	{"Single Arm Dumbbell Row", "Back", "Dumbbell"},
	// Biceps
	{"Barbell Curl", "Biceps", "Barbell"},
	{"Hammer Curls", "Biceps", "Dumbbell"},
	{"Inclined Dumbbell Curl", "Biceps", "Dumbbell"},
	{"Preacher Curl", "Biceps", "Barbell"},
	{"Dumbbell Preacher Curl", "Biceps", "Dumbbell"},
}

func d(month, day int) time.Time {
	return time.Date(2025, time.Month(month), day, 8, 0, 0, 0, time.UTC)
}

var sessionDefs = []sessionDef{
	{d(1, 21), "leg_a", 1, 0, []sg{
		{"Back Squat", 50, 8, 4, false},
		{"Deadlift", 45, 10, 3, false},
		{"Leg Press", 80, 12, 3, false},
		{"Leg Curls", 20, 12, 1, false},
		{"Leg Curls", 20, 10, 2, false},
	}},
	{d(1, 22), "push_b", 1, 0, []sg{
		{"Overhead Barbell Press", 30, 6, 2, false},
		{"Overhead Barbell Press", 30, 5, 1, true}, // last set hit failure at 5
		{"Bench Press", 15, 6, 3, false},
		{"Cable Flyes", 10, 10, 3, false},
		{"Skull Crushers", 15, 5, 1, true}, // stopped due to pain
		{"Cable Pushdown", 15, 12, 3, false},
		{"Close Grip Bench Press", 30, 8, 3, false},
	}},
	{d(1, 23), "pull_b", 1, 0, []sg{
		{"Assisted Pull Up", 40, 6, 4, false},
		{"Seated Cable Row", 35, 10, 3, false},
		{"Single Arm Dumbbell Row", 15, 10, 3, false},
		{"Inclined Dumbbell Curl", 5, 10, 3, false},
		{"Preacher Curl", 10, 12, 2, false},
		{"Preacher Curl", 15, 6, 2, false},
	}},
	{d(1, 24), "leg_b", 1, 0, []sg{
		{"Deadlift", 50, 6, 4, false},
		{"Front Squat", 40, 10, 3, false},
		{"Walking Lunges", 12.5, 10, 3, false},
		{"Leg Curls", 20, 12, 1, false},
		{"Seated Calf Raises", 40, 20, 3, false},
	}},
	{d(1, 26), "push_a", 2, 0, []sg{
		{"Bench Press", 40, 8, 4, false},
		{"Incline Dumbbell Press", 12.5, 9, 3, false},
		{"Shoulder Press", 10, 10, 3, true},
		{"Cable Pushdown", 15, 8, 1, false},
		{"Cable Pushdown", 10, 10, 2, false},
		{"Lateral Raises", 5, 12, 3, false},
		{"Cable Pushdown", 15, 15, 2, false},
	}},
	{d(1, 27), "pull_a", 2, 0, []sg{
		{"Barbell Row", 40, 6, 4, false},
		{"Lat Pulldown", 35, 10, 3, false},
		{"Face Pulls", 15, 15, 3, false},
		{"Barbell Curl", 15, 10, 3, false},
		{"Hammer Curls", 5, 10, 3, false},
		{"Rear Delt Fly", 5, 15, 2, false},
		{"Dumbbell Preacher Curl", 2.5, 60, 1, false},
	}},
	{d(1, 28), "leg_a", 2, 0, []sg{
		{"Back Squat", 50, 8, 4, false},
		{"Romanian Deadlift", 40, 10, 3, false},
		{"Leg Press", 90, 12, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Standing Calf Raises", 50, 15, 3, false},
		{"Hanging Leg Raises", 0, 12, 3, false},
	}},
	{d(1, 29), "push_b", 2, 90, []sg{
		{"Overhead Barbell Press", 30, 6, 2, false},
		{"Overhead Barbell Press", 25, 8, 2, false},
		{"Flat Dumbbell Press", 12.5, 10, 3, false},
		{"Cable Flyes", 10, 10, 3, false},
		{"Cable Pushdown", 20, 10, 2, false},
		{"Cable Pushdown", 15, 12, 2, false},
		{"Lateral Raises", 5, 15, 1, false},
		{"Lateral Raises", 7.5, 9, 1, true},
		{"Lateral Raises", 5, 12, 1, false},
		{"Diamond Push Ups", 0, 9, 1, false},
		{"Diamond Push Ups", 0, 5, 1, false},
	}},
	{d(1, 30), "pull_b", 2, 80, []sg{
		{"Assisted Pull Up", 25, 6, 4, false},
		{"Seated Cable Row", 40, 8, 3, false},
		{"Single Arm Dumbbell Row", 15, 10, 3, false},
		{"Inclined Dumbbell Curl", 5, 12, 3, false},
		{"Preacher Curl", 15, 12, 3, true},
		{"Face Pulls", 15, 15, 2, false},
	}},
	{d(2, 2), "leg_b", 3, 80, []sg{
		{"Conventional Deadlift", 60, 6, 4, false},
		{"Front Squat", 40, 10, 3, false},
		{"Walking Lunges", 12.5, 10, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Seated Calf Raises", 40, 20, 3, false},
		{"Plank", 0, 60, 3, false}, // reps = seconds
		{"Pallof Press", 20, 10, 2, false},
	}},
	{d(2, 3), "push_a", 3, 80, []sg{
		{"Bench Press", 40, 8, 4, false},
		{"Incline Dumbbell Press", 5, 10, 3, false},
		{"Shoulder Press", 7.5, 12, 3, true},
		{"Overhead Tricep Extension", 15, 10, 2, false},
		{"Cable Pushdown", 20, 10, 1, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Cable Pushdown", 15, 15, 3, false},
		{"Close Grip Push Ups", 0, 6, 2, false},
	}},
	{d(2, 4), "pull_a", 3, 0, []sg{
		{"Barbell Row", 40, 8, 4, false},
		{"Lat Pulldown", 35, 12, 3, false},
		{"Face Pulls", 20, 15, 3, false},
		{"Barbell Curl", 15, 15, 3, false},
		{"Hammer Curls", 7.5, 10, 3, true},
		{"Rear Delt Fly", 7.5, 12, 2, false},
	}},
	{d(2, 5), "leg_a", 3, 90, []sg{
		{"Back Squat", 50, 8, 4, false},
		{"Romanian Deadlift", 50, 8, 3, false},
		{"Leg Press", 90, 12, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Standing Calf Raises", 50, 15, 3, false},
		{"Hanging Leg Raises", 0, 12, 3, false},
		{"Dead Bug", 0, 12, 1, false},
		{"Dead Bug", 0, 10, 1, false},
	}},
	{d(2, 6), "push_b", 3, 75, []sg{
		{"Overhead Barbell Press", 25, 8, 4, false},
		{"Flat Dumbbell Press", 12.5, 10, 3, false},
		{"Cable Flyes", 10, 10, 3, false},
		{"Skull Crushers", 15, 10, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Close Grip Bench Press", 30, 10, 3, false},
	}},
	{d(2, 7), "pull_b", 3, 70, []sg{
		{"Assisted Pull Up", 20, 8, 4, false},
		{"Seated Cable Row", 40, 8, 3, false},
		{"Single Arm Dumbbell Row", 15, 10, 3, false},
		{"Inclined Dumbbell Curl", 5, 12, 3, false},
		{"Preacher Curl", 15, 12, 3, false},
		{"Face Pulls", 15, 15, 2, false},
	}},
	{d(2, 10), "leg_b", 4, 90, []sg{
		{"Conventional Deadlift", 60, 6, 4, false},
		{"Front Squat", 40, 10, 3, false},
		{"Walking Lunges", 12.5, 10, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Seated Calf Raises", 45, 20, 3, false},
		{"Plank", 0, 60, 3, false},
		{"Pallof Press", 20, 12, 2, false},
	}},
	{d(2, 11), "push_a", 4, 0, []sg{
		{"Bench Press", 50, 8, 4, false},
		{"Incline Dumbbell Press", 12.5, 12, 3, false},
		{"Shoulder Press", 7.5, 12, 3, false},
		{"Overhead Tricep Extension", 15, 12, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Cable Pushdown", 20, 13, 2, false},
	}},
	{d(2, 12), "pull_a", 4, 0, []sg{
		{"Barbell Row", 45, 8, 4, false},
		{"Lat Pulldown", 35, 12, 3, false},
		{"Face Pulls", 20, 15, 3, false},
		{"Barbell Curl", 15, 15, 3, false},
		{"Hammer Curls", 7.5, 10, 3, false},
	}},
	{d(2, 13), "leg_a", 4, 0, []sg{
		{"Back Squat", 55, 8, 4, false},
		{"Romanian Deadlift", 50, 8, 3, false},
		{"Leg Press", 90, 12, 3, false},
		{"Leg Curls", 50, 12, 3, false},
		{"Standing Calf Raises", 60, 15, 3, false},
		{"Hanging Leg Raises", 0, 12, 3, false},
	}},
	{d(2, 15), "push_b", 4, 0, []sg{
		{"Overhead Barbell Press", 30, 6, 4, false},
		{"Flat Dumbbell Press", 15, 8, 3, false},
		{"Cable Flyes", 12, 10, 3, false},
		{"Skull Crushers", 15, 10, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Close Grip Bench Press", 30, 10, 3, false},
	}},
	{d(2, 24), "push_a", 6, 0, []sg{
		{"Bench Press", 50, 6, 4, false},
		{"Incline Dumbbell Press", 12.5, 12, 3, false},
		{"Shoulder Press", 10, 10, 3, false},
		{"Overhead Tricep Extension", 15, 12, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Cable Pushdown", 20, 15, 1, false},
		{"Cable Pushdown", 20, 10, 1, false},
	}},
	{d(2, 25), "pull_a", 6, 0, []sg{
		{"Barbell Row", 45, 8, 4, false},
		{"Lat Pulldown", 40, 10, 3, false},
		{"Face Pulls", 20, 15, 3, false},
		{"Barbell Curl", 15, 15, 3, false},
		{"Hammer Curls", 5, 12, 2, false},
		{"Hammer Curls", 7.5, 10, 1, false},
		{"Rear Delt Fly", 5, 20, 2, false},
	}},
	{d(2, 26), "leg_a", 6, 0, []sg{
		{"Back Squat", 55, 8, 4, false},
		{"Romanian Deadlift", 50, 8, 3, false},
		{"Leg Press", 95, 12, 3, false},
		{"Leg Curls", 50, 12, 3, false},
		{"Standing Calf Raises", 50, 15, 3, false},
		{"Hanging Leg Raises", 0, 12, 3, false},
		{"Dead Bug", 0, 12, 2, false},
	}},
	{d(3, 2), "push_b", 7, 0, []sg{
		{"Overhead Barbell Press", 30, 6, 4, false},
		{"Flat Dumbbell Press", 15, 8, 3, false},
		{"Cable Flyes", 10, 10, 3, false},
		{"Skull Crushers", 15, 10, 3, false},
	}},
	{d(3, 3), "pull_b", 7, 0, []sg{
		{"Pull Up", 0, 4, 4, false},
		{"Seated Cable Row", 40, 8, 3, false},
		{"Single Arm Dumbbell Row", 15, 12, 3, false},
		{"Inclined Dumbbell Curl", 7.5, 10, 3, false},
		{"Face Pulls", 20, 15, 2, false},
		{"Preacher Curl", 15, 12, 3, false},
	}},
	{d(3, 4), "leg_b", 7, 0, []sg{
		{"Conventional Deadlift", 70, 6, 4, false},
		{"Front Squat", 50, 8, 3, false},
		{"Walking Lunges", 12.5, 10, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Seated Calf Raises", 60, 20, 3, false},
		{"Plank", 0, 60, 3, false},
		{"Pallof Press", 20, 12, 2, false},
	}},
	{d(3, 6), "push_a", 7, 0, []sg{
		{"Incline Dumbbell Press", 12.5, 12, 4, false},
		{"Bench Press", 50, 8, 3, false},
		{"Shoulder Press", 10, 10, 3, false},
		{"Overhead Tricep Extension", 15, 12, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Cable Pushdown", 20, 13, 2, false},
	}},
	{d(3, 9), "pull_a", 8, 0, []sg{
		{"Barbell Row", 45, 8, 4, false},
		{"Lat Pulldown", 40, 10, 3, false},
		{"Face Pulls", 20, 15, 3, false},
		{"Barbell Curl", 15, 15, 3, false},
		{"Rear Delt Fly", 5, 15, 2, false},
		{"Hammer Curls", 7.5, 10, 3, false},
	}},
	{d(3, 10), "leg_a", 8, 0, []sg{
		{"Back Squat", 60, 6, 4, false},
		{"Romanian Deadlift", 50, 8, 3, false},
		{"Leg Press", 95, 12, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Standing Calf Raises", 50, 15, 3, false},
		{"Hanging Leg Raises", 0, 12, 3, false},
		{"Dead Bug", 0, 12, 2, false},
	}},
	{d(3, 12), "push_b", 8, 0, []sg{
		{"Overhead Barbell Press", 30, 8, 4, false},
		{"Flat Dumbbell Press", 15, 10, 3, false},
		{"Cable Flyes", 10, 12, 3, false},
		{"Skull Crushers", 10, 12, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Close Grip Bench Press", 30, 10, 3, false},
	}},
	{d(3, 13), "leg_b", 8, 0, []sg{
		{"Conventional Deadlift", 70, 8, 4, false},
		{"Front Squat", 50, 10, 3, false},
		{"Walking Lunges", 12.5, 10, 3, false},
		{"Leg Curls", 10, 12, 3, false},
		{"Seated Calf Raises", 65, 15, 3, false},
		{"Plank", 0, 65, 3, false},
		{"Pallof Press", 20, 12, 2, false},
	}},
	{d(3, 16), "pull_b", 9, 0, []sg{
		{"Pull Up", 0, 5, 4, false},
		{"Seated Cable Row", 40, 10, 3, false},
		{"Single Arm Dumbbell Row", 15, 12, 3, false},
		{"Inclined Dumbbell Curl", 7.5, 10, 3, false},
		{"Face Pulls", 20, 15, 1, false},
		{"Face Pulls", 25, 15, 1, false},
		{"Preacher Curl", 15, 12, 3, false},
	}},
	{d(3, 17), "push_a", 9, 0, []sg{
		{"Incline Dumbbell Press", 12.5, 12, 4, false},
		{"Bench Press", 50, 6, 4, false},
		{"Shoulder Press", 10, 12, 3, false},
		{"Overhead Tricep Extension", 15, 12, 3, false},
		{"Lateral Raises", 5, 15, 3, false},
		{"Cable Pushdown", 20, 12, 2, false},
	}},
	{d(3, 19), "pull_a", 9, 0, []sg{
		{"Barbell Row", 45, 8, 4, false},
		{"Lat Pulldown", 40, 10, 3, false},
		{"Face Pulls", 20, 15, 3, false},
		{"Barbell Curl", 15, 15, 3, false},
		{"Rear Delt Fly", 5, 20, 2, false},
		{"Hammer Curls", 7.5, 10, 3, false},
	}},
}

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Clear existing seed data
	_, err = pool.Exec(ctx, "TRUNCATE exercises, programs CASCADE")
	if err != nil {
		log.Fatalf("truncate: %v", err)
	}

	// 1. Insert exercises
	exerciseIDs := make(map[string]pgtype.UUID)
	for _, e := range exerciseDefs {
		var id pgtype.UUID
		err := pool.QueryRow(ctx,
			"INSERT INTO exercises (name, muscle_group, equipment) VALUES ($1, $2, $3) RETURNING id",
			e.name, e.muscleGroup, e.equipment,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert exercise %s: %v", e.name, err)
		}
		exerciseIDs[e.name] = id
	}
	log.Printf("inserted %d exercises", len(exerciseIDs))

	// 2. Insert program
	var programID pgtype.UUID
	err = pool.QueryRow(ctx,
		"INSERT INTO programs (name, description, total_weeks) VALUES ($1, $2, $3) RETURNING id",
		"PPL A/B Split", "Push Pull Legs alternating A/B program", 9,
	).Scan(&programID)
	if err != nil {
		log.Fatalf("insert program: %v", err)
	}

	// 3. Insert program days
	type programDay struct {
		key    string
		weekNum, dayNum int
		label  string
	}
	programDays := []programDay{
		{"push_a", 1, 1, "Push A"},
		{"pull_a", 1, 2, "Pull A"},
		{"leg_a", 1, 3, "Leg A"},
		{"push_b", 2, 1, "Push B"},
		{"pull_b", 2, 2, "Pull B"},
		{"leg_b", 2, 3, "Leg B"},
	}

	dayIDs := make(map[string]pgtype.UUID)
	for _, pd := range programDays {
		var id pgtype.UUID
		err := pool.QueryRow(ctx,
			"INSERT INTO program_days (program_id, week_number, day_number, label) VALUES ($1, $2, $3, $4) RETURNING id",
			programID, pd.weekNum, pd.dayNum, pd.label,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert program day %s: %v", pd.label, err)
		}
		dayIDs[pd.key] = id
	}

	// 4. Insert sessions and exercise sets
	for _, s := range sessionDefs {
		dayID := dayIDs[s.dayKey]

		var sessionID pgtype.UUID
		err := pool.QueryRow(ctx,
			"INSERT INTO sessions (program_day_id, week_number, started_at) VALUES ($1, $2, $3) RETURNING id",
			dayID, s.weekNum, s.date,
		).Scan(&sessionID)
		if err != nil {
			log.Fatalf("insert session %s: %v", s.date.Format("2006-01-02"), err)
		}

		setNumberByExercise := make(map[pgtype.UUID]int)
		for _, setGroup := range s.sets {
			exID, ok := exerciseIDs[setGroup.exercise]
			if !ok {
				log.Fatalf("unknown exercise %q in session %s", setGroup.exercise, s.date.Format("2006-01-02"))
			}

			for i := 0; i < setGroup.numSets; i++ {
				setNumberByExercise[exID]++
				isFailure := setGroup.failure && i == setGroup.numSets-1

				_, err := pool.Exec(ctx,
					"INSERT INTO exercise_sets (session_id, exercise_id, set_number, weight_kg, reps, failure, logged_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
					sessionID, exID, setNumberByExercise[exID], setGroup.weightKg, setGroup.reps, isFailure, s.date,
				)
				if err != nil {
					log.Fatalf("insert set %s %s: %v", s.date.Format("2006-01-02"), setGroup.exercise, err)
				}
			}
		}
	}

	log.Printf("seeded %d sessions", len(sessionDefs))
}
