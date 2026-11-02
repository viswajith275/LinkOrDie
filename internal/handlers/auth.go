package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"link_or_die/internal/apperror"
	"link_or_die/internal/components"
	"link_or_die/internal/config"
	"link_or_die/internal/domains/auth"
	"link_or_die/internal/domains/user"
	"link_or_die/internal/oauth"
	"link_or_die/internal/services"
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

func (a *AuthHandler) Register(c *gin.Context) {
	components.Register(user.RegisterView{}).Render(c.Request.Context(), c.Writer)
}

func (a *AuthHandler) Login(c *gin.Context) {
	if a.adapter.Get(c, "userID") != nil {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}
	components.Login(auth.LoginView{}).Render(c.Request.Context(), c.Writer)
}

func (a *AuthHandler) LoginWithPassword(c *gin.Context) {
	var req user.LoginRequest

	if err := c.ShouldBind(&req); err != nil {
		components.Login(auth.LoginView{
			Username: req.Username,
			Error:    "Username and password are required",
		}).Render(c.Request.Context(), c.Writer)
		return
	}

	u, err := a.svc.LoginWithPassword(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, apperror.ErrInvalidCredentials) {
			components.Login(auth.LoginView{
				Username: req.Username,
				Error:    "Invalid username or password",
			}).Render(c.Request.Context(), c.Writer)
			return
		}
		internalError(c, err)
		return
	}

	a.adapter.RenewToken(c)
	a.adapter.Put(c, "userID", u.ID.String())
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (a *AuthHandler) RegisterWithPassword(c *gin.Context) {
	var req user.RegisterRequest

	if err := c.ShouldBind(&req); err != nil {
		components.Register(user.RegisterView{
			Username: req.Username,
			Email:    req.Email,
			Error:    "Input validation error: username 3-32 chars, valid email, password_len ≥ 8 chars",
		}).Render(c.Request.Context(), c.Writer)
		return
	}

	_, err := a.svc.Register(c.Request.Context(), req)
	if err != nil {
		var ve *apperror.ValidationError
		if errors.As(err, &ve) {
			components.Register(user.RegisterView{
				Username: req.Username,
				Email:    req.Email,
				Error:    ve.Message,
			}).Render(c.Request.Context(), c.Writer)
			return
		}
		internalError(c, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/login")
}

func (a *AuthHandler) Logout(c *gin.Context) {
	if err := a.adapter.Destroy(c); err != nil {
		internalError(c, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
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
