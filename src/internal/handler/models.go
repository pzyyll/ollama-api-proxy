package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"ollama-api-proxy/src/internal/dto"
	"ollama-api-proxy/src/internal/dto/ollama"
	"ollama-api-proxy/src/internal/dto/openai"
	"ollama-api-proxy/src/internal/state"

	"github.com/gin-gonic/gin"
)

func GetModels(state *state.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseUrl, err := url.Parse(state.Config.OpenAIBaseURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Invalid base URL"})
			return
		}
		apiKey := state.Config.OpenAIAPIKey

		destUrl := baseUrl.JoinPath("models").String()
		slog.Info("Fetching models from OpenAI API", "url", destUrl)

		request, _ := http.NewRequest("GET", destUrl, nil)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		// Send the request and handle the Response
		resp, err := state.HttpClient.Do(request)
		if err != nil {
			slog.Error("Failed to fetch models", "error", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to fetch models"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Error("Failed to fetch models", "status", resp.Status)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to fetch models"})
			return
		}

		var modelsResponse openai.ListModels
		if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
			slog.Error("Failed to decode models response", "error", err)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to decode models response"})
			return
		}

		var response ollama.ListResponse
		response.Models = make([]ollama.ListModelResponse, len(modelsResponse.Data))
		for i, model := range modelsResponse.Data {
			response.Models[i] = ollama.ListModelResponse{
				Name:       model.Id,
				Model:      model.Id,
				ModifiedAt: time.Unix(model.Created, 0),
				Size:       0,
				Digest:     "",
			}
		}

		c.JSON(http.StatusOK, response)
	}
}
