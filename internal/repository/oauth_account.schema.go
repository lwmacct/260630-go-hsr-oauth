package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type OauthAccountModel struct {
	bun.BaseModel `bun:"table:auth_oauth_accounts,alias:aoa"`

	ID                    string    `bun:"id,pk,type:uuid"`
	UserID                string    `bun:"user_id,type:uuid,notnull"`
	Provider              string    `bun:"provider,notnull,unique:provider_subject"`
	Subject               string    `bun:"subject,notnull,unique:provider_subject"`
	ProviderEmail         string    `bun:"provider_email,nullzero"`
	ProviderEmailVerified bool      `bun:"provider_email_verified,notnull"`
	ProviderDisplayName   string    `bun:"provider_display_name,nullzero"`
	ProviderAvatarURL     string    `bun:"provider_avatar_url,nullzero"`
	ProviderProfile       string    `bun:"provider_profile_json,nullzero"`
	CreatedAt             time.Time `bun:"created_at,notnull"`
	UpdatedAt             time.Time `bun:"updated_at,notnull"`
}

func (*OauthAccountModel) BeforeCreateTable(_ context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey("(user_id) REFERENCES users (id) ON DELETE CASCADE")
	return nil
}
