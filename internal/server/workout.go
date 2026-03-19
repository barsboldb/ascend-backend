package server

import (
  "context"
  "fmt"

  pb "github.com/barsboldb/ascend-backend/gen/workout"
)

type WorkoutServer struct {
  pb.UnimplementedWorkoutServiceServer
}

func (s *WorkoutServer) LogWorkout(ctx context.Context, req *pb.LogWorkoutRequest) (*pb.LogWorkoutResponse, error) {
  fmt.Printf("Logging workout: %s for %d minutes\n", req.Name, req.DurationMinutes)

  return &pb.LogWorkoutResponse{Id: "some-generated id"}, nil
}

func (s *WorkoutServer) GetWorkout(ctx context.Context, req *pb.GetWorkoutRequest) (*pb.GetWorkoutResponse, error) {
  return &pb.GetWorkoutResponse{
    Id: req.Id,
    Name: "Morning Run",
    DurationMinutes: 30,
  }, nil
}
