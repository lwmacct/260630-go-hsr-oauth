package oauth

import (
	"context"
	"time"

	"github.com/lwmacct/260630-go-hsr-auth/pkg/auth"
)

const (
	ProviderGitHub = "github"
	ProviderGoogle = "google"
)

type Config struct {
	Enabled         bool
	AutoRegister    bool
	CallbackBaseURL string
	Providers       []ProviderConfig
	Request         RequestFunc
	Provider        func(string) (Provider, error)
	Identity        Identity
}

type ProviderConfig struct {
	Provider string
	Label    string
}

type Options struct {
	DB      DB
	Config  Config
	FlowTTL time.Duration
}

type RequestFunc func(context.Context) (auth.SessionRequest, bool)

type Identity interface {
	UserByID(ctx context.Context, id string) (*auth.User, error)
	CreateExternalUser(ctx context.Context, input auth.ExternalUserInput) (*auth.User, error)
	CreateSession(ctx context.Context, userID string, request auth.SessionRequest) (*auth.Session, error)
}

type Provider interface {
	Name() string
	AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string
	ExchangeProfile(ctx context.Context, code string, redirectURI string, codeVerifier string) (Profile, error)
}

type Profile struct {
	Provider              string
	Subject               string
	ProviderEmail         string
	ProviderEmailVerified bool
	ProviderDisplayName   string
	ProviderAvatarURL     string
	ProviderProfile       string
}

type ClientConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}
