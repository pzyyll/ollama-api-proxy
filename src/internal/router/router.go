package router

import (
	"log/slog"
	"net/http"
	"ollama-api-proxy/src/internal/handler"
	"ollama-api-proxy/src/internal/state"

	"github.com/gin-gonic/gin"
)

func SetupRouter(engine *gin.Engine, appState *state.State) error {
	// Define routes here (e.g., engine.GET("/path", handlerFunc))

	apiRouter := engine.Group("/api")
	{
		apiRouter.GET("/version", handler.GetVersion)
		apiRouter.GET("/tags", handler.GetModels(appState))
		apiRouter.POST("/show", handler.GetModel(appState))
	}

	// OpenAI API
	v1Router := engine.Group("/v1")
	{
		v1Router.POST("/chat/completions", handler.ChatCompletion(appState))
	}

	engine.NoRoute(func(c *gin.Context) {
		slog.Info("Not Implemented", "path", c.Request.URL.Path, "method", c.Request.Method)
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not Implemented"})
	})

	return nil
}
