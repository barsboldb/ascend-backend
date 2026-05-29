package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	authpb "github.com/barsboldb/ascend-backend/gen/auth"
	programpb "github.com/barsboldb/ascend-backend/gen/program"
	sessionpb "github.com/barsboldb/ascend-backend/gen/session"
	"github.com/barsboldb/ascend-backend/internal/auth"
	"github.com/barsboldb/ascend-backend/internal/server"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.UnaryInterceptor))
	authpb.RegisterAuthServiceServer(grpcServer, server.NewAuthServer(db))
	sessionpb.RegisterSessionServiceServer(grpcServer, server.NewSessionServer(db))
	programpb.RegisterProgramServiceServer(grpcServer, server.NewProgramServer(db))
	reflection.Register(grpcServer)

	log.Println("Server running on :8888")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
