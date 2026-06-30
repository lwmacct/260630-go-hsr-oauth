package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/repository"
)

type OauthAccountService struct {
	store *repository.Store
}

func NewOauthAccountService(store *repository.Store) *OauthAccountService {
	if store == nil {
		panic("NewOauthAccountService: store is nil")
	}
	return &OauthAccountService{store: store}
}

func (s *OauthAccountService) ByProviderSubject(ctx context.Context, provider string, subject string) (*OauthAccount, error) {
	row, err := s.store.FetchOauthAccountByProviderSubject(ctx, strings.TrimSpace(provider), strings.TrimSpace(subject))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrOauthAccountNotFound
		}
		return nil, err
	}
	account := utilOauthAccount(*row)
	return &account, nil
}

func (s *OauthAccountService) Create(ctx context.Context, userID string, profile OauthProfile) (*OauthAccount, error) {
	now := time.Now().UTC()
	item := repository.OauthAccountCreate{
		UserID:                userID,
		Provider:              strings.TrimSpace(profile.Provider),
		Subject:               strings.TrimSpace(profile.Subject),
		ProviderEmail:         strings.TrimSpace(profile.ProviderEmail),
		ProviderEmailVerified: profile.ProviderEmailVerified,
		ProviderDisplayName:   strings.TrimSpace(profile.ProviderDisplayName),
		ProviderAvatarURL:     strings.TrimSpace(profile.ProviderAvatarURL),
		ProviderProfile:       profile.ProviderProfile,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if item.Provider == "" || item.Subject == "" || item.UserID == "" {
		return nil, ErrOauthAccountTaken
	}
	row, err := s.store.CreateOauthAccount(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOauthAccountTaken, err)
	}
	account := utilOauthAccount(*row)
	return &account, nil
}

func (s *OauthAccountService) DeleteForUsers(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	return s.store.DeleteOauthAccountForUsers(ctx, userIDs)
}
