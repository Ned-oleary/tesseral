package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	intermediatev1 "github.com/tesseral-labs/tesseral/internal/intermediate/gen/tesseral/intermediate/v1"
)

func (s *Service) GetGoogleOAuthRedirectURL(ctx context.Context, req *connect.Request[intermediatev1.GetGoogleOAuthRedirectURLRequest]) (*connect.Response[intermediatev1.GetGoogleOAuthRedirectURLResponse], error) {
	res, err := s.Store.GetGoogleOAuthRedirectURL(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) RedeemGoogleOAuthCode(ctx context.Context, req *connect.Request[intermediatev1.RedeemGoogleOAuthCodeRequest]) (*connect.Response[intermediatev1.RedeemGoogleOAuthCodeResponse], error) {
	res, err := s.Store.RedeemGoogleOAuthCode(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
