package core

import (
	"fmt"
	"log/slog"

	"ollma-api-proxy/src/internal/config"
	"ollma-api-proxy/src/internal/router"
	"ollma-api-proxy/src/internal/state"

	"github.com/gin-gonic/gin"
)

func InitRouterEngine(config *config.Config) *gin.Engine {
	engine := gin.New()

	engine.SetTrustedProxies(config.TrustDomains)
	engine.ForwardedByClientIP = true

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	router.SetupRouter(&state.State{
		Config: config,
		Router: engine,
	})

	return engine
}

func Run(engine *gin.Engine, config *config.Config) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	slog.Info("Starting server", "address", addr)
	if err := engine.Run(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
