package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"ollama-api-proxy/src/internal/config"
	"ollama-api-proxy/src/internal/core"
	"ollama-api-proxy/src/internal/state"
	"os"
	"testing"

	"maps"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	// Init test configuration
	envConfig := initConfig()
	models, _ := config.LoadModels("models.yaml")

	appState := &state.State{
		Config:     envConfig,
		HttpClient: &http.Client{},
		Models:     models,
	}

	testRouter = core.InitRouterEngine(appState)

	// Run the tests
	exitCode := m.Run()

	// Exit with the appropriate code
	os.Exit(exitCode)
}

func makeRequest(method, path string, body io.Reader, header map[string]string) *http.Request {
	req, _ := http.NewRequest(method, path, body)
	if header != nil {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}
	return req
}

func makeJSONRequest(method, path string, body any, header map[string]string) *http.Request {
	var payload io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		payload = bytes.NewBuffer(jsonBytes)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if header != nil {
		maps.Copy(headers, header)
	}

	return makeRequest(method, path, payload, headers)
}

func performRequest(router *gin.Engine, req *http.Request) *http.Response {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

func TestVersionAPI(t *testing.T) {
	resp := performRequest(testRouter, makeJSONRequest("GET", "/api/version", nil, nil))
	assert.Equal(t, 200, resp.StatusCode, "Expected status code 200")
}