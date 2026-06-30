package service

import (
	"errors"
	"time"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/repository"
)

type OauthAccount struct {
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

type OauthProfile struct {
	Provider              string
	Subject               string
	ProviderEmail         string
	ProviderEmailVerified bool
	ProviderDisplayName   string
	ProviderAvatarURL     string
	ProviderProfile       string
}

var (
	ErrOauthAccountRegistrationDisabled = errors.New("oauth registration disabled")
	ErrOauthAccountNotFound             = errors.New("oauth account not found")
	ErrOauthAccountTaken                = errors.New("oauth account taken")
	ErrUserDisabled                     = errors.New("user disabled")
)

func utilOauthAccount(row repository.OauthAccountRow) OauthAccount {
	return OauthAccount{
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
