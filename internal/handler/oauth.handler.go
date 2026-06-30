package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lwmacct/260630-go-hsr-auth/pkg/auth"
	"github.com/lwmacct/260630-go-hsr-oauth/internal/service"
)

type oauthHandler struct {
	config   Config
	services Services
}

func RegisterOauth(api huma.API, config Config, services Services) {
	handler := oauthHandler{config: config, services: services}
	authGroup := huma.NewGroup(api, "/auth")
	huma.Register(authGroup, huma.Operation{OperationID: "get-oauth-config", Method: http.MethodGet, Path: "/oauth/config", Tags: []string{"Oauth"}}, handler.configOutput)
	huma.Register(authGroup, huma.Operation{OperationID: "start-oauth-login", Method: http.MethodGet, Path: "/oauth/start", DefaultStatus: http.StatusFound, Tags: []string{"Oauth"}}, handler.start)
	huma.Register(authGroup, huma.Operation{OperationID: "complete-oauth-login", Method: http.MethodGet, Path: "/oauth/callback", DefaultStatus: http.StatusFound, Tags: []string{"Oauth"}}, handler.callback)
}

func (h oauthHandler) configOutput(_ context.Context, _ *struct{}) (*BodyDTO[OauthConfigDTO], error) {
	body := OauthConfigDTO{Enabled: h.config.Enabled}
	if h.config.Enabled {
		body.Providers = make([]ProviderDTO, 0, len(h.config.Providers))
		for _, provider := range h.config.Providers {
			body.Providers = append(body.Providers, ProviderDTO(provider))
		}
	}
	return &BodyDTO[OauthConfigDTO]{Body: body}, nil
}

func (h oauthHandler) start(ctx context.Context, input *OauthStartInputDTO) (*OauthStartRedirectResponseDTO, error) {
	provider, err := utilProvider(h.config, input.Provider)
	if err != nil {
		return nil, huma.Error404NotFound("oauth provider unavailable")
	}
	request, err := utilRequest(ctx, h.config)
	if err != nil {
		return nil, err
	}
	returnTo := utilSanitizeReturnTo(input.ReturnTo)
	state, flow, err := h.services.Flows.Create(ctx, provider.Name(), returnTo)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error")
	}
	redirectURI := utilOauthRedirectURI(h.config, request, provider.Name())
	return &OauthStartRedirectResponseDTO{
		Location: provider.AuthorizationURL(state, redirectURI, flow.PKCECodeVerifier, flow.Nonce),
	}, nil
}

func (h oauthHandler) callback(ctx context.Context, input *OauthCallbackInputDTO) (*OauthCallbackRedirectResponseDTO, error) {
	provider, err := utilProvider(h.config, input.Provider)
	if err != nil {
		return nil, huma.Error404NotFound("oauth provider unavailable")
	}
	request, err := utilRequest(ctx, h.config)
	if err != nil {
		return nil, err
	}
	flow, err := h.services.Flows.Consume(ctx, input.State)
	if err != nil || flow.Provider != provider.Name() {
		return nil, huma.Error401Unauthorized("invalid oauth state")
	}
	redirectURI := utilOauthRedirectURI(h.config, request, provider.Name())
	profile, err := provider.ExchangeProfile(ctx, input.Code, redirectURI, flow.PKCECodeVerifier)
	if err != nil {
		return nil, huma.Error401Unauthorized("oauth exchange failed")
	}
	user, err := h.resolveUser(ctx, profile)
	if err != nil {
		if errors.Is(err, service.ErrOauthAccountRegistrationDisabled) {
			return nil, huma.Error403Forbidden("oauth registration disabled")
		}
		if errors.Is(err, service.ErrUserDisabled) {
			return nil, huma.Error403Forbidden("user disabled")
		}
		return nil, huma.Error500InternalServerError("internal server error")
	}
	session, err := h.config.Identity.CreateSession(ctx, user.ID, request)
	if err != nil {
		return nil, huma.Error500InternalServerError("internal server error")
	}
	return &OauthCallbackRedirectResponseDTO{
		Location:  utilSanitizeReturnTo(flow.ReturnTo),
		SetCookie: session.SetCookie,
	}, nil
}

func (h oauthHandler) resolveUser(ctx context.Context, profile service.OauthProfile) (*auth.User, error) {
	if h.config.Identity == nil {
		return nil, errors.New("identity dependency unavailable")
	}
	account, err := h.services.Accounts.ByProviderSubject(ctx, profile.Provider, profile.Subject)
	if err == nil {
		user, userErr := h.config.Identity.UserByID(ctx, account.UserID)
		if userErr != nil {
			return nil, userErr
		}
		if !utilUserActive(user) {
			return nil, service.ErrUserDisabled
		}
		return user, nil
	}
	if !errors.Is(err, service.ErrOauthAccountNotFound) {
		return nil, err
	}
	if !h.config.AutoRegister {
		return nil, service.ErrOauthAccountRegistrationDisabled
	}
	user, err := h.config.Identity.CreateExternalUser(ctx, auth.ExternalUserInput{
		Username:    utilOauthUsername(profile),
		DisplayName: profile.ProviderDisplayName,
		Email:       profile.ProviderEmail,
		AvatarURL:   profile.ProviderAvatarURL,
	})
	if err != nil {
		return nil, err
	}
	if _, err := h.services.Accounts.Create(ctx, user.ID, profile); err != nil {
		return nil, err
	}
	return user, nil
}
