package handler

import (
	"context"

	"github.com/lwmacct/260630-go-hsr-auth/pkg/auth"
	"github.com/lwmacct/260630-go-hsr-oauth/internal/service"
)

type Config struct {
	Enabled         bool
	AutoRegister    bool
	CallbackBaseURL string
	Providers       []ProviderConfig
	Request         RequestAuth
	Provider        func(string) (ProviderAuth, error)
	Identity        IdentityAuth
}

type Services struct {
	Accounts *service.OauthAccountService
	Flows    *service.OauthFlowService
}

type IdentityAuth interface {
	UserByID(ctx context.Context, id int64) (*auth.User, error)
	CreateExternalUser(ctx context.Context, input auth.ExternalUserInput) (*auth.User, error)
	CreateSession(ctx context.Context, userID int64, request auth.SessionRequest) (*auth.Session, error)
}

type ProviderAuth interface {
	Name() string
	AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string
	ExchangeProfile(ctx context.Context, code string, redirectURI string, codeVerifier string) (service.OauthProfile, error)
}

type RequestAuth func(context.Context) (auth.SessionRequest, bool)

type ProviderConfig struct {
	Provider string
	Label    string
}
