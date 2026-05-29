package server

import (
	"context"
	"errors"
	"strings"

	pb "github.com/barsboldb/ascend-backend/gen/auth"
	authpkg "github.com/barsboldb/ascend-backend/internal/auth"
	"github.com/barsboldb/ascend-backend/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	db *gorm.DB
}

func NewAuthServer(db *gorm.DB) *AuthServer {
	return &AuthServer{db: db}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, status.Error(codes.InvalidArgument, "valid email is required")
	}
	if len(req.Password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	user := model.User{Email: email, PasswordHash: string(hash)}
	if strings.TrimSpace(req.Name) != "" {
		name := strings.TrimSpace(req.Name)
		user.Name = &name
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// First user inherits orphaned sessions
	var userCount int64
	s.db.Model(&model.User{}).Count(&userCount)
	if userCount == 1 {
		s.db.Model(&model.Session{}).Where("user_id IS NULL").Update("user_id", user.ID)
	}

	return s.tokenResponse(&user)
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var user model.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to look up user: %v", result.Error)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	return s.tokenResponse(&user)
}

func (s *AuthServer) Me(ctx context.Context, req *pb.MeRequest) (*pb.User, error) {
	userID, ok := ctx.Value(authpkg.UserIDKey).(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "not authenticated")
	}
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	return userToPB(&user), nil
}

func (s *AuthServer) tokenResponse(user *model.User) (*pb.AuthResponse, error) {
	token, err := authpkg.IssueToken(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to issue token: %v", err)
	}
	return &pb.AuthResponse{Token: token, User: userToPB(user)}, nil
}

func userToPB(user *model.User) *pb.User {
	name := ""
	if user.Name != nil {
		name = *user.Name
	}
	return &pb.User{Id: user.ID.String(), Email: user.Email, Name: name}
}
