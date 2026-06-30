package handler

import "time"

type HealthOutputDTO struct {
	Body HealthResponseDTO
}

type HealthResponseDTO struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
