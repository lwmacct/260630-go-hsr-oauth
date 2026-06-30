package repository

import "time"

type OauthAccountCreate struct {
	ID                    string
	UserID                string
	Provider              string
	Subject               string
	ProviderEmail         string
	ProviderEmailVerified bool
	ProviderDisplayName   string
	ProviderAvatarURL     string
	ProviderProfile       string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type OauthAccountRow struct {
	ID                    string
	UserID                string
	Provider              string
	Subject               string
	ProviderEmail         string
	ProviderEmailVerified bool
	ProviderDisplayName   string
	ProviderAvatarURL     string
	ProviderProfile       string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func utilOauthAccountRow(row OauthAccountModel) OauthAccountRow {
	return OauthAccountRow{
		ID:                    row.ID,
		UserID:                row.UserID,
		Provider:              row.Provider,
		Subject:               row.Subject,
		ProviderEmail:         row.ProviderEmail,
		ProviderEmailVerified: row.ProviderEmailVerified,
		ProviderDisplayName:   row.ProviderDisplayName,
		ProviderAvatarURL:     row.ProviderAvatarURL,
		ProviderProfile:       row.ProviderProfile,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}
