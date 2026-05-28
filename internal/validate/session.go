package validate

import (
	pb "github.com/barsboldb/ascend-backend/gen/session"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateGetSessionRequest(req *pb.GetSessionRequest) (error) {
  if req.Id == "" {
    return status.Error(codes.InvalidArgument, "id is required")
  }
  err := uuid.Validate(req.Id)
  if err != nil {
    return status.Error(codes.InvalidArgument, "invalid UUID format")
  }

  return nil
}

func ValidateCreateSessionRequest(req *pb.CreateSessionRequest) (error) {
  if req.ProgramDayId == "" {
    return status.Error(codes.InvalidArgument, "program day id is required")
  }
  err := uuid.Validate(req.ProgramDayId)
  if err != nil {
    return status.Error(codes.InvalidArgument, "invalid program day id UUID format")
  }


  return nil
}
