package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"ollama-api-proxy/src/internal/config"
	"ollama-api-proxy/src/internal/dto"
	"ollama-api-proxy/src/internal/dto/ollama"
	"ollama-api-proxy/src/internal/dto/openai"
	"ollama-api-proxy/src/internal/state"

	"github.com/gin-gonic/gin"
)

var (
	cacheModels *ollama.ListResponse
	cacheMutex  sync.Mutex
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

		func() {
			cacheMutex.Lock()
			defer cacheMutex.Unlock()
			cacheModels = &response
		}()

		c.JSON(http.StatusOK, response)
	}
}

func GetModel(state *state.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ollama.ShowRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("Invalid request body", "error", err)
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid model name"})
			return
		}

		if req.Model == "" {
			slog.Warn("Model not found", "model", req.Model)
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Model name is required"})
			return
		}

		var modelInfo *config.ModelInfo
		if state.Models != nil {
			modelInfo, _ = state.Models.GetModel(req.Model)
		}

		if modelInfo == nil {
			slog.Warn("Model not found, use default config", "model", req.Model)
			modelInfo = config.DefaultModelInfo()
		}

		c.JSON(http.StatusOK, ollama.ShowResponse{
			License:    "",
			Modelfile:  "",
			Parameters: "",
			Template:   "",
			System:     "",
			Details:    ollama.ModelDetails{
				Format: "gguf",
			},
			Messages:   []ollama.Message{},
			ModelInfo: map[string]any{
				"general.architecture": "llama",
				"llama.context_length": modelInfo.GetContextLength(),
			},
			ProjectorInfo: nil,
			Tensors:       []ollama.Tensor{},
			Capabilities:  modelInfo.GetCapabilities(),
			ModifiedAt:    time.Now(),
		})
	}
}
