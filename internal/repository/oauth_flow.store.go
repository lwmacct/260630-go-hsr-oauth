package repository

import (
	"context"
	"time"
)

func (s *Store) CreateOauthFlow(ctx context.Context, item OauthFlowCreate) (*OauthFlowRow, error) {
	row := OauthFlowModel{
		StateHash:        item.StateHash,
		Provider:         item.Provider,
		PKCECodeVerifier: item.PKCECodeVerifier,
		Nonce:            item.Nonce,
		ReturnTo:         item.ReturnTo,
		ExpiresAt:        item.ExpiresAt,
		CreatedAt:        item.CreatedAt,
	}
	if _, err := s.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return nil, err
	}
	result := utilOauthFlowRow(row)
	return &result, nil
}

func (s *Store) FetchOauthFlow(ctx context.Context, stateHash []byte) (*OauthFlowRow, error) {
	row := new(OauthFlowModel)
	if err := s.db.NewSelect().Model(row).Where("state_hash = ?", stateHash).Scan(ctx); err != nil {
		return nil, WrapNotFound(err)
	}
	result := utilOauthFlowRow(*row)
	return &result, nil
}

func (s *Store) DeleteOauthFlow(ctx context.Context, stateHash []byte) error {
	_, err := s.db.NewDelete().
		Model((*OauthFlowModel)(nil)).
		Where("state_hash = ?", stateHash).
		Exec(ctx)
	return err
}

func (s *Store) DeleteOauthFlowExpired(ctx context.Context, expiresBefore time.Time) error {
	_, err := s.db.NewDelete().
		Model((*OauthFlowModel)(nil)).
		Where("expires_at < ?", expiresBefore).
		Exec(ctx)
	return err
}
