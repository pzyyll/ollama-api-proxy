package config

import (
	"os"
	"testing"

	"ollama-api-proxy/src/internal/types/model"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("PROXY_PORT", "8080")
	os.Setenv("PROXY_HOST", "example.com")
	os.Setenv("PROXY_OPENAI_BASE_URL", "https://api.example.com/v1")
	os.Setenv("PROXY_OPENAI_API_KEY", "test-api-key")
	os.Setenv("PROXY_TRUST_DOMAINS", "example.com,localhost")
	os.Setenv("PROXY_TIMEOUT", "30s")

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "example.com", config.Host)
	assert.Equal(t, "https://api.example.com/v1", config.OpenAIBaseURL)
	assert.Equal(t, "test-api-key", config.OpenAIAPIKey)
	assert.Equal(t, "info", config.LogLevel)
	assert.Contains(t, config.TrustDomains[0], "example.com")
	assert.Contains(t, config.TrustDomains[1], "localhost")
	assert.Equal(t, 30, int(config.Timeout.Seconds()), "Timeout should be 30 seconds")
}

func TestModelsConfig(t *testing.T) {
	models, err := LoadModels("config_test.yml")
	assert.NotNil(t, models)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(models.mapModels), "There should be 2 models loaded")

	model1, err := models.GetModel("gpt-4.1")
	assert.NotNil(t, model1, "gpt-4.1 model should be loaded")
	assert.Equal(t, "gpt-4.1", model1.Name, "Model name should be gpt-4.1")
	assert.NotNil(t, model1.baseModel, "Base model should not be nil")

	model2, err := models.GetModel("gpt-4.1-mini")
	assert.NotNil(t, model2, "gpt-4.1-mini model should be loaded")
	assert.Equal(t, "gpt-4.1-mini", model2.Name, "Model name should be gpt-4.1-mini")

	capabilities := model2.GetCapabilities()
	assert.NoError(t, err, "Should not error when getting capabilities")
	assert.Equal(t, capabilities, []model.Capability{"completion", "tools"}, "Capabilities should match")
}
