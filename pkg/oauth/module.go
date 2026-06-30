package oauth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lwmacct/260630-go-hsr-auth/pkg/auth"
	"github.com/uptrace/bun"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/handler"
	"github.com/lwmacct/260630-go-hsr-oauth/internal/repository"
	"github.com/lwmacct/260630-go-hsr-oauth/internal/service"
)

type DB = bun.IDB

type Module struct {
	endpoint *handler.Endpoint
}

func New(options Options) (*Module, error) {
	if options.DB == nil {
		panic("oauth.New: DB is nil")
	}
	if options.Config.Request == nil {
		options.Config.Request = auth.RequestFromContext
	}
	store := repository.NewStore(options.DB)
	accounts := service.NewOauthAccountService(store)
	flows := service.NewOauthFlowService(store, options.FlowTTL)
	endpoint := handler.NewEndpoint(toHandlerConfig(options.Config), handler.Services{
		Accounts: accounts,
		Flows:    flows,
	})
	return &Module{endpoint: endpoint}, nil
}

func MustNew(options Options) *Module {
	module, err := New(options)
	if err != nil {
		panic(err)
	}
	return module
}

func (m *Module) Handler() http.Handler {
	return m.endpoint.Handler()
}

func (m *Module) Register(api huma.API) {
	m.endpoint.Register(api)
}

func toHandlerConfig(config Config) handler.Config {
	return handler.Config{
		Enabled:         config.Enabled,
		AutoRegister:    config.AutoRegister,
		CallbackBaseURL: config.CallbackBaseURL,
		Providers:       toHandlerProviderConfigs(config.Providers),
		Request:         handler.RequestAuth(config.Request),
		Provider:        toHandlerProvider(config.Provider),
		Identity:        handler.IdentityAuth(config.Identity),
	}
}

func toHandlerProviderConfigs(values []ProviderConfig) []handler.ProviderConfig {
	items := make([]handler.ProviderConfig, 0, len(values))
	for _, value := range values {
		items = append(items, handler.ProviderConfig{
			Provider: value.Provider,
			Label:    value.Label,
		})
	}
	return items
}

func toHandlerProvider(fn func(string) (Provider, error)) func(string) (handler.ProviderAuth, error) {
	if fn == nil {
		return nil
	}
	return func(provider string) (handler.ProviderAuth, error) {
		value, err := fn(provider)
		if err != nil {
			return nil, err
		}
		return handlerProvider{provider: value}, nil
	}
}

type handlerProvider struct {
	provider Provider
}

func (p handlerProvider) Name() string {
	return p.provider.Name()
}

func (p handlerProvider) AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string {
	return p.provider.AuthorizationURL(state, redirectURI, codeVerifier, nonce)
}

func (p handlerProvider) ExchangeProfile(ctx context.Context, code string, redirectURI string, codeVerifier string) (service.OauthProfile, error) {
	profile, err := p.provider.ExchangeProfile(ctx, code, redirectURI, codeVerifier)
	if err != nil {
		return service.OauthProfile{}, err
	}
	return toServiceProfile(profile), nil
}

var _ handler.ProviderAuth = handlerProvider{}
