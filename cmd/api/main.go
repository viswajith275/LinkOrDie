package main

import (
	"link_or_die/internals/config"
	"link_or_die/internals/handlers"
	"link_or_die/internals/middleware"
	"link_or_die/internals/oauth"
	"link_or_die/internals/services"
	"link_or_die/internals/store"
	"time"

	"github.com/39george/scs_gin_adapter"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()

	oauthCfg := config.LoadGoogleOAuthConfig()

	googleProvider, err := oauth.NewGoogleProvider(oauthCfg)

	if err != nil {
		panic(err)
	}

	db, err := store.ConnectDatabase(cfg.DatabaseURL)

	if err != nil {
		panic(err)
	}
	defer db.Close()
	// scs for session management
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db) // Background goroutine automatically cleans expired sessions
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Secure = false // Set to true in prod

	adapter := scs_gin_adapter.New(sessionManager)

	UserRepo := store.CreateUserRepo(db)
	AuthService := services.CreateAuthService(UserRepo)
	AuthHandler := handlers.CreateAuthHandler(cfg, adapter, googleProvider, AuthService)
	UserService := services.CreateUserService(UserRepo)
	DashBoardHandler := handlers.CreateDashboardHandler(cfg, adapter, UserService)

	r := gin.Default()
	r.Use(adapter.LoadAndSave)

	r.GET("/health", AuthHandler.HealthHandler)
	r.GET("/login", AuthHandler.Login)
	r.POST("/login", AuthHandler.LoginWithPassword)
	r.GET("/register", AuthHandler.Register)
	r.POST("/register", AuthHandler.RegisterWithPassword)
	r.GET("/auth/google", AuthHandler.GoogleLogin)
	r.GET("/auth/google/callback", AuthHandler.GoogleCallback)

	authed := r.Group("/")
	authed.Use(middleware.AuthRequired(adapter))
	{
		authed.GET("/dashboard", DashBoardHandler.Dashboard)
		authed.POST("/logout", AuthHandler.Logout)
	}

	r.Run(":" + cfg.Port)
}
