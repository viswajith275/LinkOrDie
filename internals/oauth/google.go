package oauth

import (
	"context"
	"fmt"
	"link_or_die/internals/domains/auth"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type GoogleProvider struct {
	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewGoogleProvider(oauthCfg *oauth2.Config) (*GoogleProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// OIDC discovery: fetch Google's public keys ONCE at startup.
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("google oidc discovery: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oauthCfg.ClientID})

	return &GoogleProvider{config: oauthCfg, verifier: verifier}, nil
}

// AuthCodeURL builds the redirect URL with PKCE and nonce.
func (g *GoogleProvider) AuthCodeURL(state, nonce, codeVerifier string) string {
	return g.config.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(codeVerifier),
		oauth2.AccessTypeOnline,
	)
}

// ExchangeAndVerify exchanges the code, verifies the ID token,
// and validates the nonce. Returns a normalized Identity.
func (g *GoogleProvider) ExchangeAndVerify(
	ctx context.Context, code, nonce, codeVerifier string,
) (*auth.Identity, error) {
	token, err := g.config.Exchange(ctx, code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := g.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("id_token verify: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, fmt.Errorf("nonce mismatch")
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("id_token claims: %w", err)
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("id_token missing sub")
	}
	if !claims.EmailVerified {
		return nil, fmt.Errorf("google email not verified")
	}

	return &auth.Identity{
		Provider:    auth.ProviderGoogle,
		ProviderUID: claims.Sub,
		Email:       claims.Email,
		Name:        claims.Name,
	}, nil
}
