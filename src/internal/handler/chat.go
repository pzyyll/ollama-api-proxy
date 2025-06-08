package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"ollama-api-proxy/src/internal/dto/newapi"
	"ollama-api-proxy/src/internal/dto/openai"
	"ollama-api-proxy/src/internal/state"

	"github.com/gin-gonic/gin"
)

func ChatCompletion(appState *state.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req newapi.GeneralOpenAIRequest
		if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
			c.AbortWithStatusJSON(http.StatusBadRequest, openai.NewError(http.StatusBadRequest, "Request body is empty"))
			return
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, openai.NewError(http.StatusBadRequest, err.Error()))
			return
		}

		if appState.Models != nil {
			if m, err := appState.Models.GetModel(req.Model); err == nil {
				req.MaxTokens = uint(m.OutputTokens)
			} else {
				req.MaxTokens = 0
			}
		} else {
			req.MaxTokens = 0
		}

		// slog.Debug("ChatCompletion request received", "max_tokens", req.MaxTokens, "model", req.Model, "MaxCompletionTokens", req.MaxCompletionTokens)

		baseUrl, err := url.Parse(appState.Config.OpenAIBaseURL)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, openai.NewError(http.StatusInternalServerError, "Invalid base URL"))
			return
		}

		payload, err := json.Marshal(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, openai.NewError(http.StatusInternalServerError, "Failed to marshal request payload"))
			return
		}

		httpRequest, err := http.NewRequest(
			http.MethodPost,
			baseUrl.JoinPath("/chat/completions").String(),
			bytes.NewBuffer(payload),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, openai.NewError(http.StatusInternalServerError, "Failed to create HTTP request"))
			return
		}

		httpRequest.Header.Set("Authorization", "Bearer "+appState.Config.OpenAIAPIKey)
		httpRequest.Header.Set("Content-Type", "application/json")

		if req.Stream {
			httpRequest.Header.Set("Accept", "text/event-stream")
			httpRequest.Header.Set("Cache-Control", "no-cache")
			httpRequest.Header.Set("Connection", "keep-alive")

			httpResponse, err := appState.HttpClient.Do(httpRequest)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, openai.NewError(http.StatusInternalServerError, "Failed to send request to OpenAI API"))
				return
			}
			defer httpResponse.Body.Close()

			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
			c.Writer.Header().Set("Transfer-Encoding", "chunked")

			c.Stream(func(w io.Writer) bool {
				if _, err := io.Copy(w, httpResponse.Body); err != nil {
					return false
				}
				return false
			})

		} else {
			httpResponse, err := appState.HttpClient.Do(httpRequest)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, openai.NewError(http.StatusInternalServerError, "Failed to send request to OpenAI API"))
				return
			}
			defer httpResponse.Body.Close()

			if httpResponse.StatusCode != http.StatusOK {
				c.AbortWithStatusJSON(httpResponse.StatusCode, openai.NewError(httpResponse.StatusCode, "OpenAI API error"))
				return
			}

			body, err := io.ReadAll(httpResponse.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, openai.NewError(http.StatusInternalServerError, "Failed to read response from OpenAI API"))
				return
			}

			c.Data(http.StatusOK, "application/json", body)
		}
	}
}
