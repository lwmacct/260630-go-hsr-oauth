package handler

import "time"

type OauthHealthOutputDTO struct {
	Body OauthHealthResponseDTO
}

type OauthHealthResponseDTO struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
