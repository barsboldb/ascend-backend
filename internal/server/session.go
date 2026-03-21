package server

import (
	"context"
	"errors"

	pb "github.com/barsboldb/ascend-backend/gen/session"
	"github.com/barsboldb/ascend-backend/internal/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type SessionServer struct {
	pb.UnimplementedSessionServiceServer
	db *gorm.DB
}

func NewSessionServer(db *gorm.DB) *SessionServer {
	return &SessionServer{db: db}
}

type ExerciseWithSets struct {
  Exercise model.Exercise
  Sets     []model.ExerciseSet
}

func (s *SessionServer) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid UUID format")
	}

	var session model.Session
	result := s.db.WithContext(ctx).
		Preload("ExerciseSets.Exercise").
		First(&session, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "session not found")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to get session: %v", result.Error)
	}

  grouped := make(map[uuid.UUID]*ExerciseWithSets)
  var order []uuid.UUID

  for _, set := range session.ExerciseSets {
    if _, ok := grouped[set.ExerciseID]; !ok {
      grouped[set.ExerciseID] = &ExerciseWithSets{Exercise: set.Exercise}
      order = append(order, set.ExerciseID)
    }
    grouped[set.ExerciseID].Sets = append(grouped[set.ExerciseID].Sets, set)
  }

  pbExercises := make([]*pb.SessionExercise, len(order))
  for i, exID := range order {
    g := grouped[exID]
    pbSets := make([]*pb.ExerciseSet, len(g.Sets))
    for j, set := range g.Sets {
      pbSets[j] = &pb.ExerciseSet{
        SetNumber: set.SetNumber,
        WeightKg:  float32(set.WeightKg),
        Reps:      set.Reps,
        Failure:   set.Failure,
      }
    }
    pbExercises[i] = &pb.SessionExercise{
      ExerciseId:   exID.String(),
      ExerciseName: g.Exercise.Name,
      Sets:         pbSets,
    }
  }

	resp := &pb.GetSessionResponse{
		Id:           session.ID.String(),
		ProgramDayId: session.ProgramDayID.String(),
		WeekNumber:   session.WeekNumber,
		StartedAt:    timestamppb.New(session.StartedAt),
		Exercises:    pbExercises,
	}
	if session.EndedAt != nil {
		resp.EndedAt = timestamppb.New(*session.EndedAt)
	}
	if session.Notes != nil {
		resp.Notes = *session.Notes
	}

	return resp, nil
}
