package server

import (
	"context"

	pb "github.com/barsboldb/ascend-backend/gen/session"
  "google.golang.org/protobuf/types/known/timestamppb"
	"github.com/barsboldb/ascend-backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SessionServer struct {
	pb.UnimplementedSessionServiceServer
	queries *db.Queries
}

func NewSessionServer(queries *db.Queries) *SessionServer {
	return &SessionServer{queries: queries}
}

func (s *SessionServer) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	var id pgtype.UUID
	if err := id.Scan(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid UUID format")
	}

	session, err := s.queries.GetSession(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "workout not found: %v", err)
	}

	return &pb.GetSessionResponse{
		Id:              session.ID.String(),
		ProgramDayId:    session.ProgramDayID.String(),
		WeekNumber:      session.WeekNumber,
    StartedAt:       timestamppb.New(session.StartedAt.Time),
    EndedAt:         timestamppb.New(session.EndedAt.Time),
	}, nil
}
