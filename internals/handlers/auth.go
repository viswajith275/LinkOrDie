package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"link_or_die/internals/components"
	"link_or_die/internals/config"
	"link_or_die/internals/domains/auth"
	"link_or_die/internals/oauth"
	"link_or_die/internals/services"
	"net/http"

	"github.com/39george/scs_gin_adapter"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	cfg     *config.Config
	adapter *scs_gin_adapter.GinAdapter
	google  oauth.Provider
	svc     *services.AuthService
}

func CreateAuthHandler(cfg *config.Config, adapter *scs_gin_adapter.GinAdapter, googleProvider oauth.Provider, svc *services.AuthService) *AuthHandler {
	return &AuthHandler{
		cfg:     cfg,
		adapter: adapter,
		google:  googleProvider,
		svc:     svc,
	}
}

func (a *AuthHandler) HealthHandler(c *gin.Context) {
	components.Landing().Render(c.Request.Context(), c.Writer)
}

func (a *AuthHandler) Login(c *gin.Context) {
	components.Login(auth.LoginView{}).Render(c.Request.Context(), c.Writer)
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func (a *AuthHandler) GoogleLogin(c *gin.Context) {
	state := randomString(32)
	nonce := randomString(32)
	verifier := oauth2.GenerateVerifier()

	// All three are single-use and bound to this browser session.
	a.adapter.Put(c, "oauth_state", state)
	a.adapter.Put(c, "oauth_nonce", nonce)
	a.adapter.Put(c, "oauth_verifier", verifier)

	c.Redirect(http.StatusFound, a.google.AuthCodeURL(state, nonce, verifier))
}

func (a *AuthHandler) GoogleCallback(c *gin.Context) {
	// Read all three before clearing
	savedState := a.adapter.GetString(c, "oauth_state")
	savedNonce := a.adapter.GetString(c, "oauth_nonce")
	savedVerifier := a.adapter.GetString(c, "oauth_verifier")

	a.adapter.Put(c, "oauth_state", "")
	a.adapter.Put(c, "oauth_nonce", "")
	a.adapter.Put(c, "oauth_verifier", "")

	if savedState == "" || c.Query("state") != savedState {
		c.String(http.StatusBadRequest, "invalid state")
		return
	}
	if savedNonce == "" || savedVerifier == "" {
		c.String(http.StatusBadRequest, "missing oauth session")
		return
	}

	// User "Deny" on Google's consent screen
	if c.Query("error") != "" {
		components.Login(auth.LoginView{Error: "Google login was cancelled."}).
			Render(c.Request.Context(), c.Writer)
		return
	}

	identity, err := a.google.ExchangeAndVerify(
		c.Request.Context(), c.Query("code"), savedNonce, savedVerifier,
	)
	if err != nil {
		components.Login(auth.LoginView{Error: "Google login failed. Try again."}).
			Render(c.Request.Context(), c.Writer)
		return
	}

	u, err := a.svc.LoginWithOAuth(c.Request.Context(), identity)
	if err != nil {
		if !HTTPrespondError(c, err) {
			components.Login(auth.LoginView{Error: "Login failed. Please try again."}).
				Render(c.Request.Context(), c.Writer)
		}
		return
	}

	a.adapter.RenewToken(c)
	a.adapter.Put(c, "userID", u.ID.String())
	c.Redirect(http.StatusSeeOther, "/dashboard")
}
