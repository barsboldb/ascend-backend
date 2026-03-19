package main

import (
  "log"
  "net"

  "google.golang.org/grpc"
  pb "github.com/barsboldb/ascend-backend/gen/workout"
  "github.com/barsboldb/ascend-backend/internal/server"
)

func main() {
  lis, err := net.Listen("tcp", ":8888")
  if err != nil {
    log.Fatalf("failed to listen: %/", err)
  }

  grpcServer := grpc.NewServer()
  pb.RegisterWorkoutServiceServer(grpcServer, &server.WorkoutServer{})

  log.Println("Server running on :8888")
  if err := grpcServer.Serve(lis); err != nil {
    log.Fatalf("failed to serve: %v", err)
  }
}
