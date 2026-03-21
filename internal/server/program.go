package server

import (
	"context"
	"errors"

	pb "github.com/barsboldb/ascend-backend/gen/program"
	"github.com/barsboldb/ascend-backend/internal/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type ProgramServer struct {
	pb.UnimplementedProgramServiceServer
	db *gorm.DB
}

func NewProgramServer(db *gorm.DB) *ProgramServer {
	return &ProgramServer{db: db}
}

func (s *ProgramServer) GetProgram(ctx context.Context, req *pb.GetProgramRequest) (*pb.GetProgramResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid UUID format")
	}

	var program model.Program
	result := s.db.WithContext(ctx).Preload("Days").First(&program, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "program not found")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to get program: %v", result.Error)
	}
  
  pbDays := make([]*pb.ProgramDay, len(program.Days))
  for i, d := range program.Days {
    pbDays[i] = &pb.ProgramDay{
      Id:         d.ID.String(),
      WeekNumber: d.WeekNumber,
      DayNumber:  d.DayNumber,
      Label:      d.Label,
    }
  }

	resp := &pb.GetProgramResponse{
		Id:        program.ID.String(),
		Name:      program.Name,
		CreatedAt: timestamppb.New(program.CreatedAt),
    Days:      pbDays,
	}
	if program.Description != nil {
		resp.Description = *program.Description
	}
	if program.TotalWeeks != nil {
		resp.TotalWeeks = *program.TotalWeeks
	}

	return resp, nil
}

func (s *ProgramServer) GetProgramDay(ctx context.Context, req *pb.GetProgramDayRequest) (*pb.GetProgramDayResponse, error) {
  if req.Id == "" {
    return nil, status.Error(codes.InvalidArgument, "id is required")
  }

  id, err := uuid.Parse(req.Id)
  if err != nil {
    return nil, status.Error(codes.InvalidArgument, "invalid UUID format")
  }

  var programDay model.ProgramDay
  result := s.db.WithContext(ctx).
    Preload("Exercises.Exercise").
    First(&programDay, id)

  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, status.Error(codes.NotFound, "program day not found")
  }
  if result.Error != nil {
    return nil, status.Errorf(codes.Internal, "failed to get program day: %v", result.Error)
  }

  pbExercises := make([]*pb.ProgramExercise, len(programDay.Exercises))
  for i, e := range programDay.Exercises {
    var weightIncrement float32

    if e.WeightIncrement != nil {
      weightIncrement = float32(*e.WeightIncrement)
    }

    pbExercises[i] = &pb.ProgramExercise{
      Id:               e.ID.String(),
      Name:             e.Exercise.Name,
      MuscleGroup:     *e.Exercise.MuscleGroup,
      Sets:             e.Sets,
      RepMin:           e.RepMin,
      RepMax:           e.RepMax,
      IsAmrap:          e.IsAmrap,
      IsTimed:          e.IsTimed,
      WeightIncrement:  weightIncrement,
    }
  }

  resp := &pb.GetProgramDayResponse{
    Id:         programDay.ID.String(),
    WeekNumber: programDay.WeekNumber,
    DayNumber:  programDay.DayNumber,
    Label:      programDay.Label,
    Exercises:  pbExercises,
  }

  return resp, nil
}
