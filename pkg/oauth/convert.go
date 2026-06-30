package oauth

import "github.com/lwmacct/260630-go-hsr-oauth/internal/service"

func toServiceProfile(value Profile) service.OauthProfile {
	return service.OauthProfile{
		Provider:              value.Provider,
		Subject:               value.Subject,
		ProviderEmail:         value.ProviderEmail,
		ProviderEmailVerified: value.ProviderEmailVerified,
		ProviderDisplayName:   value.ProviderDisplayName,
		ProviderAvatarURL:     value.ProviderAvatarURL,
		ProviderProfile:       value.ProviderProfile,
	}
}

func fromServiceProfile(value service.OauthProfile) Profile {
	return Profile{
		Provider:              value.Provider,
		Subject:               value.Subject,
		ProviderEmail:         value.ProviderEmail,
		ProviderEmailVerified: value.ProviderEmailVerified,
		ProviderDisplayName:   value.ProviderDisplayName,
		ProviderAvatarURL:     value.ProviderAvatarURL,
		ProviderProfile:       value.ProviderProfile,
	}
}
