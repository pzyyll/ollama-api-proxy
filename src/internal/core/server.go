package core

import (
	"fmt"
	"log/slog"

	"ollama-api-proxy/src/internal/config"
	"ollama-api-proxy/src/internal/router"
	"ollama-api-proxy/src/internal/state"

	"github.com/gin-gonic/gin"
)

func InitRouterEngine(appState *state.State) *gin.Engine {
	engine := gin.New()

	engine.SetTrustedProxies(appState.Config.TrustDomains)
	engine.ForwardedByClientIP = true

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	router.SetupRouter(engine, appState)

	return engine
}

func Run(engine *gin.Engine, config *config.Config) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	slog.Info("Starting server", "address", config.Host, "port", config.Port)

	if err := engine.Run(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
