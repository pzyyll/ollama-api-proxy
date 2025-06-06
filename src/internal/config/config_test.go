package config

import (
	"os"
	"testing"

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