package auth

import (
	"context"
	"errors"
	sso "github.com/dojyaaa-n/protos/gen/go/sso"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso/internal/services/auth"
	"sso/utils/storage"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int64,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

type serverAPI struct {
	sso.UnimplementedAuthServer
	auth Auth
}

type LoginRequestValidation struct {
	Email    string `validate:"required"`
	Password string `validate:"required"`
	AppId    int64  `validate:"required,min=1"`
}

type RegisterRequestValidation struct {
	Email    string `validate:"required"`
	Password string `validate:"required"`
}

type IsAdminRequestValidation struct {
	UserId int64 `validate:"required,min=1"`
}

func RegisterServer(gRPC *grpc.Server, auth *auth.Auth) {
	sso.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *sso.LoginRequest,
) (*sso.LoginResponse, error) {
	request := LoginRequestValidation{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		AppId:    req.GetAppId(),
	}

	err := validator.New().Struct(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Request validation failed")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &sso.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *sso.RegisterRequest,
) (*sso.RegisterResponse, error) {
	request := RegisterRequestValidation{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	err := validator.New().Struct(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Request validation failed")
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &sso.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *sso.IsAdminRequest,
) (*sso.IsAdminResponse, error) {
	request := IsAdminRequestValidation{
		UserId: req.GetUserId(),
	}

	err := validator.New().Struct(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Request validation failed")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.AlreadyExists, "User not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &sso.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}
