package clientgrpc

import (
	"context"

	clientv1 "github.com/deeploop-ai/fleet/genproto/client/v1"
	sharedv1 "github.com/deeploop-ai/fleet/genproto/shared/v1"
	"github.com/deeploop-ai/fleet/internal/app/client"
	"github.com/deeploop-ai/fleet/internal/pkg/contexts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AccountService struct {
	clientv1.UnimplementedAccountServiceServer
	account *client.Account
}

func NewAccountService(account *client.Account) *AccountService {
	return &AccountService{account: account}
}

func (s *AccountService) SignUp(ctx context.Context, req *clientv1.SignUpRequest) (*clientv1.SignUpResponse, error) {
	user, tokens, _, err := s.account.SignUp(ctx, client.SignUpCommand{
		ProjectID: req.GetProjectId(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		Name:      req.GetName(),
	})
	if err != nil {
		return nil, err
	}
	return &clientv1.SignUpResponse{
		Account: mapUser(user),
		Tokens:  mapTokens(tokens),
	}, nil
}

func (s *AccountService) SignIn(ctx context.Context, req *clientv1.SignInRequest) (*clientv1.SignInResponse, error) {
	user, tokens, _, err := s.account.SignIn(ctx, client.SignInCommand{
		ProjectID: req.GetProjectId(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &clientv1.SignInResponse{
		Account: mapUser(user),
		Tokens:  mapTokens(tokens),
	}, nil
}

func (s *AccountService) SignOut(ctx context.Context, _ *clientv1.SignOutRequest) (*sharedv1.Empty, error) {
	if _, ok := contexts.Principal(ctx); !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if err := s.account.SignOut(ctx); err != nil {
		return nil, err
	}
	return &sharedv1.Empty{}, nil
}

func (s *AccountService) Me(ctx context.Context, _ *clientv1.MeRequest) (*clientv1.Account, error) {
	user, err := s.account.Me(ctx)
	if err != nil {
		return nil, err
	}
	return mapUser(user), nil
}

func mapUser(u *client.User) *clientv1.Account {
	if u == nil {
		return nil
	}
	return &clientv1.Account{
		Id:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
	}
}

func mapTokens(t *client.TokenBundle) *clientv1.TokenBundle {
	if t == nil {
		return nil
	}
	return &clientv1.TokenBundle{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    t.ExpiresAt,
	}
}
