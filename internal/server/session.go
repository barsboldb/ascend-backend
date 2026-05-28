package server

import (
	"context"
	"errors"

	pb "github.com/barsboldb/ascend-backend/gen/session"
	"github.com/barsboldb/ascend-backend/internal/mapper"
	"github.com/barsboldb/ascend-backend/internal/model"
	"github.com/barsboldb/ascend-backend/internal/validate"

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

func (s *SessionServer) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.SessionWithExercises, error) {
  if err := validate.ValidateGetSessionRequest(req); err != nil {
    return nil, err
  }

  id := uuid.MustParse(req.Id)
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

  resp := mapper.SessionToPB(&session)

	if session.EndedAt != nil {
		resp.EndedAt = timestamppb.New(*session.EndedAt)
	}
	if session.Notes != nil {
		resp.Notes = *session.Notes
	}

	return resp, nil
}

func (s *SessionServer) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.SessionWithExercises, error) {
  if err := validate.ValidateCreateSessionRequest(req); err != nil {
    return nil, err
  }

  session := mapper.PBToSession(req)

  result := s.db.Create(session)
  if result.Error != nil {
    return nil, status.Errorf(codes.Internal, "failed to create session: %v", result.Error)
  }

  return mapper.SessionToPB(session), nil;
}

func (s *SessionServer) ListRecentSessions(ctx context.Context, req *pb.ListRecentSessionsRequest) (*pb.ListRecentSessionsResponse, error) {
  limit := int(req.Limit)
  if limit <= 0 {
    limit = 10
  }
  if limit > 100 {
    limit = 100
  }

  var sessions []model.Session
  result := s.db.WithContext(ctx).
    Preload("ExerciseSets.Exercise").
    Order("started_at DESC").
    Limit(limit).
    Find(&sessions)
  if result.Error != nil {
    return nil, status.Errorf(codes.Internal, "failed to list sessions: %v", result.Error)
  }

  resp := &pb.ListRecentSessionsResponse{Sessions: make([]*pb.SessionWithExercises, len(sessions))}
  for i := range sessions {
    pbSession := mapper.SessionToPB(&sessions[i])
    if sessions[i].EndedAt != nil {
      pbSession.EndedAt = timestamppb.New(*sessions[i].EndedAt)
    }
    if sessions[i].Notes != nil {
      pbSession.Notes = *sessions[i].Notes
    }
    resp.Sessions[i] = pbSession
  }
  return resp, nil
}

func (s *SessionServer) GetLastSetsForExercise(ctx context.Context, req *pb.GetLastSetsForExerciseRequest) (*pb.GetLastSetsForExerciseResponse, error) {
  if req.ExerciseId == "" {
    return nil, status.Error(codes.InvalidArgument, "exercise_id is required")
  }
  exerciseId, err := uuid.Parse(req.ExerciseId)
  if err != nil {
    return nil, status.Error(codes.InvalidArgument, "invalid exercise_id UUID")
  }

  var latestSet model.ExerciseSet
  result := s.db.WithContext(ctx).
    Where("exercise_id = ?", exerciseId).
    Order("logged_at DESC").
    First(&latestSet)

  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return &pb.GetLastSetsForExerciseResponse{ExerciseId: req.ExerciseId, Found: false}, nil
  }
  if result.Error != nil {
    return nil, status.Errorf(codes.Internal, "failed to find last set: %v", result.Error)
  }

  var sets []model.ExerciseSet
  if err := s.db.WithContext(ctx).
    Preload("Exercise").
    Where("session_id = ? AND exercise_id = ?", latestSet.SessionID, exerciseId).
    Order("set_number ASC").
    Find(&sets).Error; err != nil {
    return nil, status.Errorf(codes.Internal, "failed to load sets: %v", err)
  }

  var session model.Session
  if err := s.db.WithContext(ctx).First(&session, latestSet.SessionID).Error; err != nil {
    return nil, status.Errorf(codes.Internal, "failed to load session: %v", err)
  }

  pbSets := make([]*pb.ExerciseSet, len(sets))
  var exerciseName string
  for i, set := range sets {
    exerciseName = set.Exercise.Name
    pbSets[i] = &pb.ExerciseSet{
      ExerciseId:   set.ExerciseID.String(),
      ExerciseName: set.Exercise.Name,
      SetNumber:    set.SetNumber,
      WeightKg:     float32(set.WeightKg),
      Reps:         set.Reps,
      Failure:      set.Failure,
      Rpe:          set.Rpe,
    }
  }

  return &pb.GetLastSetsForExerciseResponse{
    ExerciseId:   req.ExerciseId,
    ExerciseName: exerciseName,
    Sets:         pbSets,
    PerformedAt:  timestamppb.New(session.StartedAt),
    Found:        true,
  }, nil
}
