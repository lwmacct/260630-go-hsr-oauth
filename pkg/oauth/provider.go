package oauth

import (
	"context"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/infra/oauthclient"
)

func NewProvider(name string, config ClientConfig) (Provider, error) {
	provider, err := oauthclient.New(name, oauthclient.ProviderConfig{
		Enabled:      config.Enabled,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       config.Scopes,
		AuthURL:      config.AuthURL,
		TokenURL:     config.TokenURL,
		UserInfoURL:  config.UserInfoURL,
	})
	if err != nil {
		return nil, err
	}
	return clientProvider{provider: provider}, nil
}

type clientProvider struct {
	provider *oauthclient.Provider
}

func (p clientProvider) Name() string {
	return p.provider.Name()
}

func (p clientProvider) AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string {
	return p.provider.AuthorizationURL(state, redirectURI, codeVerifier, nonce)
}

func (p clientProvider) ExchangeProfile(ctx context.Context, code string, redirectURI string, codeVerifier string) (Profile, error) {
	profile, err := p.provider.ExchangeProfile(ctx, oauthclient.TokenRequest{
		Code:         code,
		RedirectURI:  redirectURI,
		CodeVerifier: codeVerifier,
	})
	if err != nil {
		return Profile{}, err
	}
	return fromServiceProfile(profile), nil
}
