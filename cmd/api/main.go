package main

import (
	"link_or_die/internals/config"
	"link_or_die/internals/handlers"
	"link_or_die/internals/store"
	"time"

	gin_adapter "github.com/39george/scs_gin_adapter"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.Load()
	db, err := store.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db) // Background goroutine automatically cleans expired sessions
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Secure = false // Set to true in prod

	adapter := gin_adapter.New(sessionManager)

	r := gin.Default()
	r.Use(adapter.LoadAndSave)

	r.GET("/health", handlers.HealthHandler)

	r.Run(":" + cfg.Port)
}
