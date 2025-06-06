package state

import (
	"net/http"

	"ollama-api-proxy/src/internal/config"

	"github.com/gin-gonic/gin"
)

type State struct {
	Config     *config.Config
	Router     *gin.Engine
	HttpClient *http.Client
}
