package router

import (
	"ollma-api-proxy/src/internal/handler"
	"ollma-api-proxy/src/internal/state"
)

func SetupRouter(appState *state.State) error {
	engine := appState.Router

	// Define routes here (e.g., engine.GET("/path", handlerFunc))

	apiRouter := engine.Group("/api")
	{
		apiRouter.GET("/version", handler.GetVersion)
	}

	return nil
}
