package config

import (
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func LoadGoogleOAuthConfig() *oauth2.Config {
	// setting up oauth config
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
