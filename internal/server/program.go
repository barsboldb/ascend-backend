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

	resp := &pb.GetProgramResponse{
		Id:        program.ID.String(),
		Name:      program.Name,
		CreatedAt: timestamppb.New(program.CreatedAt),
	}
	if program.Description != nil {
		resp.Description = *program.Description
	}
	if program.TotalWeeks != nil {
		resp.TotalWeeks = *program.TotalWeeks
	}

	return resp, nil
}
