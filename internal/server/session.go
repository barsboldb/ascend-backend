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
