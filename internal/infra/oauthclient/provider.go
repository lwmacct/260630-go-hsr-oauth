package oauthclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/service"
)

type ProviderConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type Provider struct {
	name   string
	cfg    ProviderConfig
	client *http.Client
}

type TokenRequest struct {
	Code         string
	RedirectURI  string
	CodeVerifier string
}

func New(name string, cfg ProviderConfig) (*Provider, error) {
	name = strings.TrimSpace(name)
	if name == "" || !cfg.Enabled || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.AuthURL == "" || cfg.TokenURL == "" || cfg.UserInfoURL == "" {
		return nil, errors.New("oauth provider disabled or incomplete")
	}
	return &Provider{name: name, cfg: cfg, client: http.DefaultClient}, nil
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string {
	values := url.Values{}
	values.Set("client_id", p.cfg.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	values.Set("state", state)
	values.Set("scope", strings.Join(p.cfg.Scopes, " "))
	values.Set("code_challenge", codeChallenge(codeVerifier))
	values.Set("code_challenge_method", "S256")
	if p.name == ProviderGoogle {
		values.Set("nonce", nonce)
	}
	u, _ := url.Parse(p.cfg.AuthURL)
	u.RawQuery = values.Encode()
	return u.String()
}

func (p *Provider) ExchangeProfile(ctx context.Context, req TokenRequest) (service.OauthProfile, error) {
	accessToken, err := p.exchangeToken(ctx, req)
	if err != nil {
		return service.OauthProfile{}, err
	}
	switch p.name {
	case ProviderGitHub:
		return p.githubProfile(ctx, accessToken)
	case ProviderGoogle:
		return p.googleProfile(ctx, accessToken)
	default:
		return service.OauthProfile{}, errors.New("unsupported oauth provider")
	}
}

func (p *Provider) exchangeToken(ctx context.Context, req TokenRequest) (string, error) {
	values := url.Values{}
	values.Set("client_id", p.cfg.ClientID)
	values.Set("client_secret", p.cfg.ClientSecret)
	values.Set("code", req.Code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", req.RedirectURI)
	values.Set("code_verifier", req.CodeVerifier)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth token exchange failed: %s", resp.Status)
	}
	var token struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return "", err
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("oauth token exchange failed: %s %s", token.Error, token.Description)
	}
	return token.AccessToken, nil
}

func (p *Provider) githubProfile(ctx context.Context, accessToken string) (service.OauthProfile, error) {
	body, err := p.getJSON(ctx, p.cfg.UserInfoURL, accessToken)
	if err != nil {
		return service.OauthProfile{}, err
	}
	var data struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return service.OauthProfile{}, err
	}
	email := data.Email
	emailVerified := email != ""
	if email == "" {
		email, emailVerified = p.githubPrimaryEmail(ctx, accessToken)
	}
	displayName := data.Name
	if displayName == "" {
		displayName = data.Login
	}
	return service.OauthProfile{
		Provider:              ProviderGitHub,
		Subject:               strconv.FormatInt(data.ID, 10),
		ProviderEmail:         email,
		ProviderEmailVerified: emailVerified,
		ProviderDisplayName:   displayName,
		ProviderAvatarURL:     data.AvatarURL,
		ProviderProfile:       string(body),
	}, nil
}

func (p *Provider) githubPrimaryEmail(ctx context.Context, accessToken string) (string, bool) {
	endpoint := "https://api.github.com/user/emails"
	if strings.HasSuffix(p.cfg.UserInfoURL, "/user") {
		endpoint = strings.TrimSuffix(p.cfg.UserInfoURL, "/user") + "/user/emails"
	}
	body, err := p.getJSON(ctx, endpoint, accessToken)
	if err != nil {
		return "", false
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", false
	}
	for _, item := range emails {
		if item.Primary {
			return item.Email, item.Verified
		}
	}
	return "", false
}

func (p *Provider) googleProfile(ctx context.Context, accessToken string) (service.OauthProfile, error) {
	body, err := p.getJSON(ctx, p.cfg.UserInfoURL, accessToken)
	if err != nil {
		return service.OauthProfile{}, err
	}
	var data struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return service.OauthProfile{}, err
	}
	return service.OauthProfile{
		Provider:              ProviderGoogle,
		Subject:               data.Sub,
		ProviderEmail:         data.Email,
		ProviderEmailVerified: data.EmailVerified,
		ProviderDisplayName:   data.Name,
		ProviderAvatarURL:     data.Picture,
		ProviderProfile:       string(body),
	}, nil
}

func (p *Provider) getJSON(ctx context.Context, endpoint string, accessToken string) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var body bytes.Buffer
	if _, err := io.Copy(&body, io.LimitReader(resp.Body, 1<<20)); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oauth profile request failed: %s", resp.Status)
	}
	return body.Bytes(), nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
