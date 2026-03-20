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

	pbSets := make([]*pb.ExerciseSet, len(session.ExerciseSets))
	for i, set := range session.ExerciseSets {
		pbSets[i] = &pb.ExerciseSet{
			ExerciseId:   set.ExerciseID.String(),
			ExerciseName: set.Exercise.Name,
			SetNumber:    set.SetNumber,
			WeightKg:     float32(set.WeightKg),
			Reps:         set.Reps,
			Failure:      set.Failure,
		}
	}

	resp := &pb.GetSessionResponse{
		Id:           session.ID.String(),
		ProgramDayId: session.ProgramDayID.String(),
		WeekNumber:   session.WeekNumber,
		StartedAt:    timestamppb.New(session.StartedAt),
		ExerciseSets: pbSets,
	}
	if session.EndedAt != nil {
		resp.EndedAt = timestamppb.New(*session.EndedAt)
	}
	if session.Notes != nil {
		resp.Notes = *session.Notes
	}

	return resp, nil
}
