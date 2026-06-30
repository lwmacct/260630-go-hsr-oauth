package service

import (
	"crypto/sha256"
	"time"

	"github.com/lwmacct/260630-go-hsr-oauth/internal/repository"
)

type OauthFlow struct {
	Provider         string
	PKCECodeVerifier string
	Nonce            string
	ReturnTo         string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

func utilOauthFlow(row repository.OauthFlowRow) OauthFlow {
	return OauthFlow{
		Provider:         row.Provider,
		PKCECodeVerifier: row.PKCECodeVerifier,
		Nonce:            row.Nonce,
		ReturnTo:         row.ReturnTo,
		ExpiresAt:        row.ExpiresAt,
		CreatedAt:        row.CreatedAt,
	}
}

func utilOauthFlowStateHash(value string) []byte {
	hash := sha256.Sum256([]byte(value))
	return hash[:]
}
