package oauth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lwmacct/260630-go-hsr-auth/pkg/auth"
	"github.com/lwmacct/260630-go-hsr-oauth/pkg/oauth"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestModuleOAuthStart(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer func() { _ = db.Close() }()
	if err := auth.ApplySchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := oauth.ApplySchema(ctx, db); err != nil {
		t.Fatal(err)
	}

	identity := mustAuthModule(t, db)
	module, err := oauth.New(oauth.Options{
		DB: db,
		Config: oauth.Config{
			Enabled:      true,
			AutoRegister: true,
			Providers: []oauth.ProviderConfig{
				{Provider: oauth.ProviderGitHub, Label: "GitHub"},
			},
			Identity: identity,
			Provider: func(provider string) (oauth.Provider, error) {
				return testProvider{name: provider}, nil
			},
		},
		FlowTTL: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

	handler := withRequestContext(module.Handler())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/oauth/start?provider=github&returnTo=%2F%23%2Fconsole", nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "state=") || !strings.Contains(location, "redirect_uri=") {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestModuleOAuthCallbackAutoRegistersAndCreatesSession(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer func() { _ = db.Close() }()
	if err := auth.ApplySchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := oauth.ApplySchema(ctx, db); err != nil {
		t.Fatal(err)
	}

	identity := mustAuthModule(t, db)
	profile := oauth.Profile{
		Provider:              oauth.ProviderGitHub,
		Subject:               "42",
		ProviderEmail:         "octo@example.test",
		ProviderEmailVerified: true,
		ProviderDisplayName:   "Octo Cat",
		ProviderAvatarURL:     "https://example.test/avatar.png",
		ProviderProfile:       `{"id":42}`,
	}
	module, err := oauth.New(oauth.Options{
		DB: db,
		Config: oauth.Config{
			Enabled:      true,
			AutoRegister: true,
			Providers: []oauth.ProviderConfig{
				{Provider: oauth.ProviderGitHub, Label: "GitHub"},
			},
			Identity: identity,
			Provider: func(provider string) (oauth.Provider, error) {
				return testProvider{name: provider, profile: profile}, nil
			},
		},
		FlowTTL: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

	oauthHandler := withRequestContext(module.Handler())
	start := httptest.NewRecorder()
	oauthHandler.ServeHTTP(start, httptest.NewRequest(http.MethodGet, "/auth/oauth/start?provider=github&returnTo=%2F%23%2Fconsole", nil))
	if start.Code != http.StatusFound {
		t.Fatalf("start status = %d, body = %s", start.Code, start.Body.String())
	}
	state := mustQueryValue(t, start.Header().Get("Location"), "state")

	callback := httptest.NewRecorder()
	callbackURL := "/auth/oauth/callback?provider=github&code=ok&state=" + url.QueryEscape(state)
	oauthHandler.ServeHTTP(callback, httptest.NewRequest(http.MethodGet, callbackURL, nil))
	if callback.Code != http.StatusFound {
		t.Fatalf("callback status = %d, body = %s", callback.Code, callback.Body.String())
	}
	if callback.Header().Get("Location") != "/#/console" {
		t.Fatalf("unexpected return location: %s", callback.Header().Get("Location"))
	}
	cookies := callback.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("missing session cookie, headers = %#v", callback.Header())
	}
	cookie := cookies[0]
	if cookie.Name != "web_session" || cookie.Value == "" {
		t.Fatalf("unexpected session cookie: %#v", cookie)
	}

	me := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(cookie)
	withRequestContext(identity.Handler()).ServeHTTP(me, meReq)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", me.Code, me.Body.String())
	}
	var session authSessionBody
	if err := json.Unmarshal(me.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if !session.Authenticated || session.User == nil || session.User.Username != "octo" {
		t.Fatalf("unexpected session body: %#v", session)
	}

	count, err := db.NewSelect().Model((*oauthAccountRow)(nil)).Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("oauth account count = %d", count)
	}
}

func openTestDB(t *testing.T) *bun.DB {
	t.Helper()

	sqlDB, err := sql.Open(sqliteshim.ShimName, "file:oauth-module-test?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	return bun.NewDB(sqlDB, sqlitedialect.New())
}

func mustAuthModule(t *testing.T, db *bun.DB) *auth.Module {
	t.Helper()
	module, err := auth.New(auth.Options{
		DB: db,
		Config: auth.Config{
			Local: auth.LocalConfig{
				LoginEnabled:        true,
				RegistrationEnabled: true,
			},
			ChallengeProvider: passChallengeProvider{},
		},
		SessionTTL: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	return module
}

func withRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := auth.SessionRequest{
			IP:         "127.0.0.1",
			Scheme:     "http",
			Host:       r.Host,
			UserAgent:  r.UserAgent(),
			Method:     r.Method,
			Path:       r.URL.Path,
			RemoteAddr: r.RemoteAddr,
		}
		next.ServeHTTP(w, r.WithContext(auth.ContextWithRequest(r.Context(), request)))
	})
}

type testProvider struct {
	name    string
	profile oauth.Profile
}

func (p testProvider) Name() string {
	return p.name
}

func (p testProvider) AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string {
	values := url.Values{}
	values.Set("state", state)
	values.Set("redirect_uri", redirectURI)
	values.Set("code_verifier", codeVerifier)
	return "https://example.test/oauth?" + values.Encode()
}

func (p testProvider) ExchangeProfile(context.Context, string, string, string) (oauth.Profile, error) {
	return p.profile, nil
}

func mustQueryValue(t *testing.T, rawURL string, key string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	value := parsed.Query().Get(key)
	if value == "" {
		t.Fatalf("missing %q in %s", key, rawURL)
	}
	return value
}

type oauthAccountRow struct {
	bun.BaseModel `bun:"table:auth_oauth_accounts"`
}

type authSessionBody struct {
	Authenticated bool          `json:"authenticated"`
	User          *authUserBody `json:"user"`
}

type authUserBody struct {
	Username string `json:"username"`
}

type passChallengeProvider struct{}

func (passChallengeProvider) Name() string {
	return "pass"
}

func (passChallengeProvider) PublicConfig() auth.ChallengePublicConfig {
	return auth.ChallengePublicConfig{Provider: "pass"}
}

func (passChallengeProvider) Create(context.Context, auth.ChallengeInput) (*auth.Challenge, error) {
	return &auth.Challenge{Provider: "pass", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (passChallengeProvider) Verify(context.Context, auth.ChallengeAnswer, auth.ChallengeInput) error {
	return nil
}
