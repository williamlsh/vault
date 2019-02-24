package vault

import (
	"context"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

// NewHTTPServer makes a new Vault HTTP service.
func NewHTTPServer(ctx context.Context, endpoints Endpoints) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/hash", httptransport.NewServer(endpoints.HashEndpoint,
		decodeHashRequest,
		encodeResponse,
	))
	mux.Handle("/validate", httptransport.NewServer(
		endpoints.ValidateEndpoint,
		decodeValidateRequest,
		encodeResponse,
	))
	return mux
}
