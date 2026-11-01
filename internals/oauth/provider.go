package oauth

import (
	"context"
	"link_or_die/internals/domains/auth"
)

type Provider interface {
	// AuthCodeURL builds the redirect URL with PKCE and nonce.
	AuthCodeURL(state, nonce, codeVerifier string) string
	// ExchangeAndVerify exchanges the code, verifies the ID token, and validates the nonce.
	ExchangeAndVerify(ctx context.Context, code, nonce, codeVerifier string) (*auth.Identity, error)
}
