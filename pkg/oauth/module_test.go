package oauth_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
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
	name string
}

func (p testProvider) Name() string {
	return p.name
}

func (p testProvider) AuthorizationURL(state string, redirectURI string, codeVerifier string, nonce string) string {
	return "https://example.test/oauth?state=" + state + "&redirect_uri=" + redirectURI + "&code_verifier=" + codeVerifier
}

func (p testProvider) ExchangeProfile(context.Context, string, string, string) (oauth.Profile, error) {
	return oauth.Profile{}, nil
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
