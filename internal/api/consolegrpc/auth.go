package consolegrpc

import (
	"context"

	consolev1 "github.com/deeploop-ai/orionid/genproto/console/v1"
	"github.com/deeploop-ai/orionid/internal/app/console"
)

type AuthService struct {
	consolev1.UnimplementedConsoleAuthServiceServer
	auth *console.Auth
}

func NewAuthService(auth *console.Auth) *AuthService {
	return &AuthService{auth: auth}
}

func (s *AuthService) SignIn(ctx context.Context, req *consolev1.SignInRequest) (*consolev1.SignInResponse, error) {
	token, expiresAt, err := s.auth.SignIn(ctx, console.SignInCommand{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &consolev1.SignInResponse{AccessToken: token, ExpiresAt: expiresAt}, nil
}
