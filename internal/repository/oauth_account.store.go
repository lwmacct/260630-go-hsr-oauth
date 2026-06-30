package repository

import (
	"context"

	"github.com/uptrace/bun"
)

func (s *Store) CreateOauthAccount(ctx context.Context, item OauthAccountCreate) (*OauthAccountRow, error) {
	row := OauthAccountModel{
		UserID:                item.UserID,
		Provider:              item.Provider,
		Subject:               item.Subject,
		ProviderEmail:         item.ProviderEmail,
		ProviderEmailVerified: item.ProviderEmailVerified,
		ProviderDisplayName:   item.ProviderDisplayName,
		ProviderAvatarURL:     item.ProviderAvatarURL,
		ProviderProfile:       item.ProviderProfile,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
	if _, err := s.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return nil, err
	}
	result := utilOauthAccountRow(row)
	return &result, nil
}

func (s *Store) FetchOauthAccountByProviderSubject(ctx context.Context, provider string, subject string) (*OauthAccountRow, error) {
	row := new(OauthAccountModel)
	if err := s.db.NewSelect().
		Model(row).
		Where("provider = ?", provider).
		Where("subject = ?", subject).
		Scan(ctx); err != nil {
		return nil, WrapNotFound(err)
	}
	result := utilOauthAccountRow(*row)
	return &result, nil
}

func (s *Store) DeleteOauthAccountForUsers(ctx context.Context, userIDs []int64) error {
	_, err := s.db.NewDelete().
		Model((*OauthAccountModel)(nil)).
		Where("user_id IN (?)", bun.List(userIDs)).
		Exec(ctx)
	return err
}
