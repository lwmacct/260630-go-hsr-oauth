package oauth

import (
	"context"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/repository"
	"github.com/lwmacct/260630-go-hsr-shared/pkg/schema"
	"github.com/uptrace/bun"
)

func Models() []any {
	return []any{
		(*repository.OauthAccountModel)(nil),
		(*repository.OauthFlowModel)(nil),
	}
}

func ApplySchema(ctx context.Context, db *bun.DB) error {
	return schema.Apply(ctx, db, Models()...)
}

func ResetSchema(ctx context.Context, db *bun.DB) error {
	return schema.Reset(ctx, db, []string{"auth_oauth_accounts", "auth_oauth_flows"}, Models()...)
}
