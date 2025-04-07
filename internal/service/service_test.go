package service

import (
	"UsersService/internal/config"
	"UsersService/internal/repository"
	"UsersService/mocks"
	"UsersService/protos/gen"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestRegister(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockJWT := new(mocks.JWTClient)
	mockLogger := zap.NewNop().Sugar()

	service := NewAuthService(config.Config{}, mockRepository, mockLogger, mockJWT)

	req := &gen.RegisterRequest{
		Username:  "testuser",
		Email:     "test@gmail.com",
		Password:  "testpassword",
		FirstName: "testname",
		LastName:  "testlastname",
	}

	mockRepository.On("RegisterUser", context.Background(), mock.Anything).Return(req.Username, nil)
	mockRepository.On("CheckUserExists", context.Background(), req.GetUsername(), req.GetEmail()).Return(false, nil)

	resp, err := service.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, &gen.RegisterResponse{Message: req.Username}, resp)

	mockRepository.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockJWT := new(mocks.JWTClient)
	mockLogger := zap.NewNop().Sugar()
	service := NewAuthService(config.Config{}, mockRepository, mockLogger, mockJWT)

	username := "testuser"
	password := "testpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &repository.User{
		Username: username,
		HashPass: string(hashedPassword),
		Role:     "user",
	}

	mockRepository.On("GetUserByUsername", context.Background(), mock.Anything).Return(user, nil)
	mockRepository.On("UpdateLoginTime", context.Background(), mock.Anything).Return(nil)
	mockJWT.On("NewJWT", mock.Anything).Return("token", nil)

	resp, err := service.Login(context.Background(), &gen.LoginRequest{
		Username: username,
		Password: password,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, &gen.LoginResponse{Token: resp.Token}, resp)

	mockRepository.AssertExpectations(t)
	mockJWT.AssertExpectations(t)

}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockJWT := new(mocks.JWTClient)
	mockLogger := zap.NewNop().Sugar()
	service := NewAuthService(config.Config{}, mockRepository, mockLogger, mockJWT)
	mockRepository.On("GetUserByUsername", context.Background(), mock.Anything).Return(nil, errors.New("unable to get user by username"))

	resp, err := service.Login(context.Background(), &gen.LoginRequest{
		Username: "",
		Password: "",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
	mockRepository.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockJWT := new(mocks.JWTClient)
	mockLogger := zap.NewNop().Sugar()
	service := NewAuthService(config.Config{}, mockRepository, mockLogger, mockJWT)

	username := "testuser"
	correctPassword := "testpassword"
	wrongPassword := "wrongPassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	user := &repository.User{
		Username: username,
		HashPass: string(hashedPassword),
	}
	mockRepository.On("GetUserByUsername", context.Background(), mock.Anything).Return(user, nil)
	resp, err := service.Login(context.Background(), &gen.LoginRequest{
		Username: username,
		Password: wrongPassword,
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	mockRepository.AssertExpectations(t)
}

func TestLogin_TokenGenerationErr(t *testing.T) {
	mockRepository := new(mocks.Repository)
	mockJWT := new(mocks.JWTClient)
	mockLogger := zap.NewNop().Sugar()
	service := NewAuthService(config.Config{}, mockRepository, mockLogger, mockJWT)

	username := "testuser"
	password := "testpassword"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &repository.User{
		Username: username,
		HashPass: string(hashedPassword),
		Role:     "user",
	}
	mockRepository.On("GetUserByUsername", context.Background(), mock.Anything).Return(user, nil)
	mockRepository.On("UpdateLoginTime", mock.Anything, username).Return(nil)
	mockJWT.On("NewJWT", mock.Anything).Return("", errors.New("Error creating JWT token"))
	resp, err := service.Login(context.Background(), &gen.LoginRequest{
		Username: username,
		Password: password,
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))

	mockRepository.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}
