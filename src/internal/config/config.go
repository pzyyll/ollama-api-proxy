package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Port          int           `koanf:"port" validate:"min=1,max=65535"`
	Host          string        `koanf:"host" validate:"hostname|ip"`
	GinMode       string        `koanf:"gin_mode" validate:"oneof=debug release test"`
	OpenAIBaseURL string        `koanf:"openai_base_url" validate:"url"`
	OpenAIAPIKey  string        `koanf:"openai_api_key"`
	LogLevel      string        `koanf:"log_level" validate:"oneof=debug info warn error"`
	TrustDomains  []string      `koanf:"trust_domains" validate:"dive,hostname|ip"`
	Timeout       time.Duration `koanf:"timeout" validate:"gte=0"`
}

func Default() *Config {
	return &Config{
		Port:          11434,
		Host:          "0.0.0.0",
		GinMode:       "debug",
		OpenAIBaseURL: "https://api.openai.com/v1",
		OpenAIAPIKey:  "",
		LogLevel:      "info",
		TrustDomains:  []string{"localhost", "127.0.0.1", "::1"},
		Timeout:       5 * time.Minute, // Default timeout of 5 minutes
	}
}

func LoadConfig() (*Config, error) {
	config := Default()

	envPrefix := "PROXY_"
	envSeparator := "_"

	k := koanf.New(envSeparator)
	if err := k.Load(env.ProviderWithValue(envPrefix, "_", func(k string, v string) (string, any) {
		key := strings.ToLower(
			strings.TrimPrefix(k, envPrefix),
		)

		if strings.Contains(v, ",") {
			var values []string
			for val := range strings.SplitSeq(v, ",") {
				values = append(values, strings.TrimSpace(val))
			}
			return key, values
		}

		return key, v
	}), nil); err != nil {
		return nil, err
	}

	if err := k.UnmarshalWithConf("", config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true}); err != nil {
		return nil, err
	}

	if err := validator.New().Struct(config); err != nil {
		var errorMsgs []string
		for _, fe := range err.(validator.ValidationErrors) {
			errorMsgs = append(errorMsgs, fmt.Sprintf("'%s': require '%s %s' (value: '%v')",
				fe.Namespace(), fe.Tag(), fe.Param(), fe.Value()))
		}
		return nil, fmt.Errorf("configuration validation failed:\n - %s",
			strings.Join(errorMsgs, "\n - "))
	}

	return config, nil
}
