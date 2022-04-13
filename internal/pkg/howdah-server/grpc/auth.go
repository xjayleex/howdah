package howdah_server

import (
	"context"
	"encoding/base64"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"howdah/pb"
	"strings"
)

type AuthService struct {
	adminStore AdminStore
}

func NewAuthService (store AdminStore) (*AuthService, error) {
	return &AuthService{
		adminStore: store,
	}, nil
}

func (a *AuthService) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	_, err := a.Authenticate(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Authentication error.")
	}

	// Todo : jwt needed?
	return &pb.SignInResponse{Token: "dummy token"}, nil
}

func (a *AuthService) Authenticate(ctx context.Context) (context.Context, error) {
	return a.tryBasicAuth(ctx)
}

func (a *AuthService) tryBasicAuth(ctx context.Context) (context.Context, error) {
	auth, err := extractHeader(ctx, "authorization")
	if err != nil {
		return ctx, err
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return ctx, status.Error(codes.Unauthenticated, `missing "Basic " prefix in "Authorization" header`)
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return ctx, status.Error(codes.Unauthenticated, `invalid base64 in header`)
	}

	cs := string(c)
	i := strings.IndexByte(cs, ':')
	if i < 0 {
		return ctx, status.Error(codes.Unauthenticated, `invalid basic auth format`)
	}

	user, password := cs[:i], cs[i+1:]

	found, err := a.adminStore.find(user)
	if err != nil {
		return ctx, status.Error(codes.Unauthenticated, `invalid authentication info`)
	}

	if found.password != password {
		return ctx, status.Error(codes.Unauthenticated, `password mismatched`)
	}

	return ctx, nil
}

type admin struct {
	id string
	password string
}

type AdminStore interface {
	find(id string) (admin, error)
}


func extractHeader(ctx context.Context, header string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no headers in request")
	}

	headers, ok := md[header]
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no header in request")
	}

	if len(headers) != 1 {
		return "", status.Error(codes.Unauthenticated, "more than 1 header in request")
	}

	return headers[0], nil
}
