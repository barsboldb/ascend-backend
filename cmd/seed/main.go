package main

import (
	"log"
	"os"
	"time"

	"github.com/barsboldb/ascend-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	// Clear existing seed data
	db.Exec("TRUNCATE exercises, programs CASCADE")

	// 1. Insert exercises
	exerciseIDs := make(map[string]uuid.UUID)
	for _, e := range exerciseDefs {
		mg := e.muscleGroup
		eq := e.equipment
		exercise := model.Exercise{
			Name:        e.name,
			MuscleGroup: &mg,
			Equipment:   &eq,
		}
		if err := db.Create(&exercise).Error; err != nil {
			log.Fatalf("insert exercise %s: %v", e.name, err)
		}
		exerciseIDs[e.name] = exercise.ID
	}
	log.Printf("inserted %d exercises", len(exerciseIDs))

	// 2. Insert program with days
	desc := "12-Week PPL Muscle Building Program for Ectomorphs"
	weeks := int32(12)
	program := model.Program{
		Name:        "PPL A/B Split",
		Description: &desc,
		TotalWeeks:  &weeks,
		Days: []model.ProgramDay{
			{WeekNumber: 1, DayNumber: 1, Label: "Push A"},
			{WeekNumber: 1, DayNumber: 2, Label: "Pull A"},
			{WeekNumber: 1, DayNumber: 3, Label: "Leg A"},
			{WeekNumber: 1, DayNumber: 4, Label: "Push B"},
			{WeekNumber: 1, DayNumber: 5, Label: "Pull B"},
			{WeekNumber: 1, DayNumber: 6, Label: "Leg B"},
			{WeekNumber: 1, DayNumber: 7, Label: "Rest"},
		},
	}
	if err := db.Create(&program).Error; err != nil {
		log.Fatalf("insert program: %v", err)
	}

	dayIDs := make(map[string]uuid.UUID)
	for _, day := range program.Days {
		dayIDs[day.Label] = day.ID
	}

	// 3. Insert program exercises
	type pe struct {
		day      string
		exercise string
		position int32
		sets     int32
		repMin   int32
		repMax   int32
		isAmrap  bool
		isTimed  bool
	}

	programExercises := []pe{
		// Push A — Bench Press Focus
		{"Push A", "Bench Press", 1, 4, 6, 8, false, false},
		{"Push A", "Incline Dumbbell Press", 2, 3, 8, 10, false, false},
		{"Push A", "Shoulder Press", 3, 3, 10, 12, false, false},
		{"Push A", "Overhead Tricep Extension", 4, 3, 10, 12, false, false},
		{"Push A", "Lateral Raises", 5, 3, 12, 15, false, false},
		{"Push A", "Cable Pushdown", 6, 2, 12, 15, false, false},
		// Pull A — Row Focus
		{"Pull A", "Barbell Row", 1, 4, 6, 8, false, false},
		{"Pull A", "Lat Pulldown", 2, 3, 8, 10, false, false},
		{"Pull A", "Face Pulls", 3, 3, 12, 15, false, false},
		{"Pull A", "Barbell Curl", 4, 3, 8, 10, false, false},
		{"Pull A", "Hammer Curls", 5, 3, 10, 12, false, false},
		{"Pull A", "Rear Delt Fly", 6, 2, 12, 15, false, false},
		// Leg A — Squat Focus
		{"Leg A", "Back Squat", 1, 4, 6, 8, false, false},
		{"Leg A", "Romanian Deadlift", 2, 3, 8, 10, false, false},
		{"Leg A", "Leg Press", 3, 3, 10, 12, false, false},
		{"Leg A", "Leg Curls", 4, 3, 10, 12, false, false},
		{"Leg A", "Standing Calf Raises", 5, 3, 12, 15, false, false},
		{"Leg A", "Hanging Leg Raises", 6, 3, 10, 15, false, false},
		{"Leg A", "Dead Bug", 7, 2, 10, 10, false, false},
		// Push B — Overhead Press Focus
		{"Push B", "Overhead Barbell Press", 1, 4, 6, 8, false, false},
		{"Push B", "Flat Dumbbell Press", 2, 3, 8, 10, false, false},
		{"Push B", "Cable Flyes", 3, 3, 10, 12, false, false},
		{"Push B", "Skull Crushers", 4, 3, 10, 12, false, false},
		{"Push B", "Lateral Raises", 5, 3, 12, 15, false, false},
		{"Push B", "Close Grip Bench Press", 6, 2, 8, 10, false, false},
		// Pull B — Vertical Pull Focus
		{"Pull B", "Pull Up", 1, 4, 0, 0, true, false},
		{"Pull B", "Seated Cable Row", 2, 3, 8, 10, false, false},
		{"Pull B", "Single Arm Dumbbell Row", 3, 3, 10, 12, false, false},
		{"Pull B", "Inclined Dumbbell Curl", 4, 3, 10, 12, false, false},
		{"Pull B", "Preacher Curl", 5, 3, 10, 12, false, false},
		{"Pull B", "Face Pulls", 6, 2, 15, 15, false, false},
		// Leg B — Deadlift Focus
		{"Leg B", "Conventional Deadlift", 1, 4, 5, 6, false, false},
		{"Leg B", "Front Squat", 2, 3, 8, 10, false, false},
		{"Leg B", "Walking Lunges", 3, 3, 10, 10, false, false},
		{"Leg B", "Leg Curls", 4, 3, 10, 12, false, false},
		{"Leg B", "Seated Calf Raises", 5, 3, 15, 20, false, false},
		{"Leg B", "Plank", 6, 3, 30, 60, false, true},
		{"Leg B", "Pallof Press", 7, 2, 10, 10, false, false},
	}

	for _, p := range programExercises {
		exID, ok := exerciseIDs[p.exercise]
		if !ok {
			log.Fatalf("unknown exercise %q in program day %s", p.exercise, p.day)
		}
		pe := model.ProgramExercise{
			ProgramDayID: dayIDs[p.day],
			ExerciseID:   exID,
			Position:     p.position,
			Sets:         p.sets,
			RepMin:       p.repMin,
			RepMax:       p.repMax,
			IsAmrap:      p.isAmrap,
			IsTimed:      p.isTimed,
		}
		if err := db.Create(&pe).Error; err != nil {
			log.Fatalf("insert program exercise %s %s: %v", p.day, p.exercise, err)
		}
	}
	log.Printf("inserted %d program exercises", len(programExercises))

	labelByKey := map[string]string{
		"push_a": "Push A", "pull_a": "Pull A", "leg_a": "Leg A",
		"push_b": "Push B", "pull_b": "Pull B", "leg_b": "Leg B",
	}

	// 3. Insert sessions and exercise sets
	for _, s := range sessionDefs {
		dayID := dayIDs[labelByKey[s.dayKey]]

		session := model.Session{
			ProgramDayID: dayID,
			WeekNumber:   int32(s.weekNum),
			StartedAt:    s.date,
		}

		setNumberByExercise := make(map[uuid.UUID]int32)
		for _, setGroup := range s.sets {
			exID, ok := exerciseIDs[setGroup.exercise]
			if !ok {
				log.Fatalf("unknown exercise %q in session %s", setGroup.exercise, s.date.Format("2006-01-02"))
			}
			for i := 0; i < setGroup.numSets; i++ {
				setNumberByExercise[exID]++
				session.ExerciseSets = append(session.ExerciseSets, model.ExerciseSet{
					ExerciseID: exID,
					SetNumber:  setNumberByExercise[exID],
					WeightKg:   setGroup.weightKg,
					Reps:       int32(setGroup.reps),
					Failure:    setGroup.failure && i == setGroup.numSets-1,
					LoggedAt:   s.date,
				})
			}
		}

		if err := db.Create(&session).Error; err != nil {
			log.Fatalf("insert session %s: %v", s.date.Format("2006-01-02"), err)
		}
	}

	log.Printf("seeded %d sessions", len(sessionDefs))
}
