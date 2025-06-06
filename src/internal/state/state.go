package state

import (
	"ollma-api-proxy/src/internal/config"

	"github.com/gin-gonic/gin"
)

type State struct {
	Config *config.Config
	Router *gin.Engine
}
