package handler

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lwmacct/260630-go-hsr-auth/pkg/auth"
	"github.com/lwmacct/260630-go-hsr-oauth/internal/service"
)

func utilOauthRedirectURI(config Config, request auth.SessionRequest, provider string) string {
	base := strings.TrimRight(config.CallbackBaseURL, "/")
	if base == "" {
		scheme := request.Scheme
		if scheme == "" {
			scheme = "http"
		}
		host := request.Host
		if host == "" {
			host = request.RemoteAddr
		}
		base = scheme + "://" + host
	}
	values := url.Values{}
	values.Set("provider", provider)
	return base + "/api/auth/oauth/callback?" + values.Encode()
}

func utilSanitizeReturnTo(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/#/console"
	}
	if strings.HasPrefix(value, "#/") {
		return "/" + value
	}
	if strings.HasPrefix(value, "/#/") || strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return value
	}
	return "/#/console"
}

func utilOauthUsername(profile service.OauthProfile) string {
	if profile.ProviderEmail != "" {
		if name, _, ok := strings.Cut(profile.ProviderEmail, "@"); ok {
			return name
		}
	}
	if profile.ProviderDisplayName != "" {
		return profile.ProviderDisplayName
	}
	return profile.Provider + "-" + profile.Subject
}

func utilRequest(ctx context.Context, config Config) (auth.SessionRequest, error) {
	if config.Request == nil {
		return auth.SessionRequest{}, huma.Error400BadRequest("invalid request source")
	}
	request, ok := config.Request(ctx)
	if !ok {
		return auth.SessionRequest{}, huma.Error400BadRequest("invalid request source")
	}
	return request, nil
}

func utilProvider(config Config, provider string) (ProviderAuth, error) {
	if !config.Enabled || config.Provider == nil {
		return nil, errors.New("oauth disabled")
	}
	return config.Provider(strings.ToLower(strings.TrimSpace(provider)))
}

func utilUserActive(user *auth.User) bool {
	return user != nil && user.Status != auth.UserStatusDisabled && user.DisabledAt == nil
}
