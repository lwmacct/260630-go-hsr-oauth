package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type Endpoint struct {
	config   Config
	services Services
}

func NewEndpoint(config Config, services Services) *Endpoint {
	return &Endpoint{config: config, services: services}
}

func (e *Endpoint) Handler() http.Handler {
	mux := http.NewServeMux()
	api := humago.New(mux, utilHTTPConfig())
	e.Register(api)
	return mux
}

func (e *Endpoint) Register(api huma.API) {
	RegisterOauth(api, e.config, e.services)
}
