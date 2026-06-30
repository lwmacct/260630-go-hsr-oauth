package handler

type OauthConfigDTO struct {
	Enabled   bool          `json:"enabled"`
	Providers []ProviderDTO `json:"providers"`
}

type ProviderDTO struct {
	Provider string `json:"provider"`
	Label    string `json:"label"`
}

type OauthStartInputDTO struct {
	Provider string `query:"provider"`
	ReturnTo string `query:"returnTo"`
}

type OauthCallbackInputDTO struct {
	Provider string `query:"provider"`
	Code     string `query:"code"`
	State    string `query:"state"`
}

type OauthStartRedirectResponseDTO struct {
	Location string `header:"Location"`
}

type OauthCallbackRedirectResponseDTO struct {
	Location  string `header:"Location"`
	SetCookie string `header:"Set-Cookie"`
}
