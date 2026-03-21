package model

import (
	"time"

	"github.com/google/uuid"
)

type Program struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string     `gorm:"not null"`
	Description *string
	TotalWeeks  *int32
	CreatedAt   time.Time
	Days        []ProgramDay `gorm:"foreignKey:ProgramID"`
}

type ProgramDay struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProgramID  uuid.UUID `gorm:"type:uuid;not null"`
	WeekNumber int32     `gorm:"not null"`
	DayNumber  int32     `gorm:"not null"`
	Label      string    `gorm:"not null"`
	Exercises  []ProgramExercise `gorm:"foreignKey:ProgramDayID"`
}

type Exercise struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"not null"`
	MuscleGroup *string
	Equipment   *string
}

type ProgramExercise struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProgramDayID    uuid.UUID `gorm:"type:uuid;not null"`
	ExerciseID      uuid.UUID `gorm:"type:uuid;not null"`
	Position        int32     `gorm:"not null"`
	Sets            int32     `gorm:"not null"`
	RepMin          int32     `gorm:"not null"`
	RepMax          int32     `gorm:"not null"`
	WeightIncrement *float64
	IsAmrap         bool      `gorm:"default:false"`
	IsTimed         bool      `gorm:"default:false"`
	Exercise        Exercise  `gorm:"foreignKey:ExerciseID"`
}

type Session struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProgramDayID uuid.UUID  `gorm:"type:uuid;not null"`
	WeekNumber   int32      `gorm:"not null"`
	StartedAt    time.Time  `gorm:"not null"`
	EndedAt      *time.Time
	Notes        *string
	ExerciseSets []ExerciseSet `gorm:"foreignKey:SessionID"`
}

type ExerciseSet struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SessionID  uuid.UUID `gorm:"type:uuid;not null"`
	ExerciseID uuid.UUID `gorm:"type:uuid;not null"`
	SetNumber  int32     `gorm:"not null"`
	WeightKg   float64   `gorm:"not null"`
	Reps       int32     `gorm:"not null"`
	Failure    bool      `gorm:"default:false"`
	LoggedAt   time.Time
	Exercise   Exercise `gorm:"foreignKey:ExerciseID"`
}
