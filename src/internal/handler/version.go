package handler

import (
	"net/http"

	"ollma-api-proxy/src/internal/constants"

	"github.com/gin-gonic/gin"
)

func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": constants.OllamaAPIVersion})
}
