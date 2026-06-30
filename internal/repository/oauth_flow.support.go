package repository

import "time"

type OauthFlowCreate struct {
	StateHash        []byte
	Provider         string
	PKCECodeVerifier string
	Nonce            string
	ReturnTo         string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

type OauthFlowRow struct {
	StateHash        []byte
	Provider         string
	PKCECodeVerifier string
	Nonce            string
	ReturnTo         string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

func utilOauthFlowRow(row OauthFlowModel) OauthFlowRow {
	return OauthFlowRow{
		StateHash:        row.StateHash,
		Provider:         row.Provider,
		PKCECodeVerifier: row.PKCECodeVerifier,
		Nonce:            row.Nonce,
		ReturnTo:         row.ReturnTo,
		ExpiresAt:        row.ExpiresAt,
		CreatedAt:        row.CreatedAt,
	}
}
