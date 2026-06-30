package service

import (
	"context"
	"errors"
	"time"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/repository"
	"github.com/lwmacct/260630-go-hsr-shared/pkg/token"
)

type OauthFlowService struct {
	store *repository.Store
	ttl   time.Duration
}

func NewOauthFlowService(store *repository.Store, ttl time.Duration) *OauthFlowService {
	if store == nil {
		panic("NewOauthFlowService: store is nil")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &OauthFlowService{store: store, ttl: ttl}
}

func (s *OauthFlowService) Create(ctx context.Context, provider string, returnTo string) (state string, flow OauthFlow, err error) {
	state, err = token.NewWithPrefix("oauth_state")
	if err != nil {
		return "", OauthFlow{}, err
	}
	codeVerifier, err := token.NewBase62(64)
	if err != nil {
		return "", OauthFlow{}, err
	}
	nonce, err := token.NewWithPrefix("nonce")
	if err != nil {
		return "", OauthFlow{}, err
	}
	now := time.Now().UTC()
	row, err := s.store.CreateOauthFlow(ctx, repository.OauthFlowCreate{
		StateHash:        utilOauthFlowStateHash(state),
		Provider:         provider,
		PKCECodeVerifier: codeVerifier,
		Nonce:            nonce,
		ReturnTo:         returnTo,
		ExpiresAt:        now.Add(s.ttl),
		CreatedAt:        now,
	})
	if err != nil {
		return "", OauthFlow{}, err
	}
	return state, utilOauthFlow(*row), nil
}

func (s *OauthFlowService) Consume(ctx context.Context, state string) (*OauthFlow, error) {
	stateHash := utilOauthFlowStateHash(state)
	row, err := s.store.FetchOauthFlow(ctx, stateHash)
	if err != nil {
		return nil, err
	}
	_ = s.store.DeleteOauthFlow(ctx, stateHash)
	if time.Now().UTC().After(row.ExpiresAt) {
		return nil, repository.ErrNotFound
	}
	flow := utilOauthFlow(*row)
	return &flow, nil
}

func (s *OauthFlowService) Cleanup(ctx context.Context) error {
	err := s.store.DeleteOauthFlowExpired(ctx, time.Now().UTC())
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	return err
}
