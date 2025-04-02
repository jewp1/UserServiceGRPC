package service

import (
	"UsersService/internal/config"
	"UsersService/internal/repository"
	"UsersService/pkg/jwt"
	"UsersService/protos/gen"
	"context"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type authService struct {
	cfg  config.Config
	repo repository.Repository
	log  *zap.SugaredLogger
	jwt  jwt.JWTClient
	gen.UnimplementedAuthServiceServer
}

func NewAuthService(
	cfg config.Config, repo repository.Repository, log *zap.SugaredLogger, jwt jwt.JWTClient) gen.AuthServiceServer {
	return &authService{
		cfg:  cfg,
		repo: repo,
		log:  log,
		jwt:  jwt,
	}
}

func (s *authService) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.RegisterResponse, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" || req.GetEmail() == "" {
		s.log.Errorf("Username and password are required")
		return nil, status.Error(codes.InvalidArgument, "username or password is empty")
	}

	exists, err := s.repo.CheckUserExists(ctx, req.GetUsername(), req.GetEmail())
	if err != nil {
		s.log.Errorf("Username or Email is invalid")
		return nil, status.Error(codes.Internal, "error checking user")
	}

	if exists {
		s.log.Errorf("User: %s already exists", req.GetUsername())
		return nil, status.Error(codes.AlreadyExists, "User already exists")
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)

	if err != nil {
		s.log.Error("failed to hash password", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "failed to generate hash")
	}
	req.Password = string(hashPass)
	s.log.Infof("started registering a user...")

	resp, err := s.repo.RegisterUser(ctx, &repository.User{
		Email:       req.GetEmail(),
		Username:    req.GetUsername(),
		HashPass:    req.GetPassword(),
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		LastLoginAt: time.Time{},
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	})
	if err != nil {
		s.log.Errorf("failed to register user: %v", err)
		return nil, status.Error(codes.Internal, "failed to register user")
	}
	s.log.Infof("User registered successfully: %s", req.Username)
	return &gen.RegisterResponse{Message: resp}, nil
}

func (s *authService) Login(ctx context.Context, req *gen.LoginRequest) (*gen.LoginResponse, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		s.log.Errorf("Username and password are required")
		return nil, status.Error(codes.InvalidArgument, "username or password is empty")
	}
	user, err := s.repo.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		s.log.Errorf("User not found: %s", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	comparHashs := bcrypt.CompareHashAndPassword([]byte(user.HashPass), []byte(req.GetPassword()))
	if comparHashs != nil {
		s.log.Errorf("invalid password for user: %s", user.Username)
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	err = s.repo.UpdateLoginTime(ctx, req.GetUsername())
	if err != nil {
		s.log.Errorf("failed to update user login time: %v", err)
		return nil, status.Error(codes.Internal, "failed to update user login time")
	}

	s.log.Infof("passwords match, a key will be generated for username: %s", user.Username)

	token, err := s.jwt.NewJWT(jwt.CreateTokenParams{
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		s.log.Errorf("failed to generate token: %v", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	s.log.Infof("token generated")
	return &gen.LoginResponse{Token: token}, nil
}
