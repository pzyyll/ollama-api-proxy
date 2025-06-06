package core

import (
	"fmt"
	"log/slog"
	"net/http"

	"ollama-api-proxy/src/internal/config"
	"ollama-api-proxy/src/internal/router"
	"ollama-api-proxy/src/internal/state"

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
		HttpClient: &http.Client{
			Timeout: config.Timeout,
		},
	})

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
