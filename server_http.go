package vault

import (
	"context"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

// NewHTTPServer makes a new Vault HTTP service.
func NewHTTPServer(ctx context.Context, endpoints Endpoints) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/hash", httptransport.NewServer(ctx, endpoints.HashEndpoint,
		decodeHashRequest,
		encodeResponse,
	))
	mux.Handle("/validate", httptransport.NewServer(ctx,
		endpoints.ValidateEndpoint,
		decodeValidateRequest,
		encodeResponse,
	))
	return mux
}
