package config

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ollama-api-proxy/src/internal/types/model"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
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

// BaseModel defines the structure for a base model configuration.

type BaseModelConfig struct {
	Capabilities []model.Capability `koanf:"capabilities,omitempty" validate:"dive,oneof=completion tools vision thinking insert"`
	InputTokens  int                `koanf:"input_tokens,omitempty"`
	OutputTokens int                `koanf:"output_tokens,omitempty"`
}

type BaseModel struct {
	Name            string `koanf:"name"`
	BaseModelConfig `koanf:"config"`
}

// ModelInfo defines the structure for a specific model configuration.
type ModelInfo struct {
	Name            string  `koanf:"name"`
	Base            *string `koanf:"base,omitempty"`
	BaseModelConfig `koanf:"config"`
	baseModel       *BaseModel `koanf:"-"`
}

func (m *ModelInfo) GetInputTokens() int {
	if m.InputTokens != 0 {
		return m.InputTokens
	}
	if m.baseModel != nil && m.baseModel.InputTokens != 0 {
		return m.baseModel.InputTokens
	}
	return 0
}

func (m *ModelInfo) GetOutputTokens() int {
	if m.OutputTokens != 0 {
		return m.OutputTokens
	}
	if m.baseModel != nil && m.baseModel.OutputTokens != 0 {
		return m.baseModel.OutputTokens
	}
	return 0
}

func (m *ModelInfo) GetContextLength() int {
	return m.GetInputTokens() + m.GetOutputTokens()
}

func (m *ModelInfo) GetCapabilities() []model.Capability {
	if m.Capabilities != nil {
		return m.Capabilities
	}
	if m.baseModel != nil && m.baseModel.Capabilities != nil {
		return m.baseModel.Capabilities
	}
	return []model.Capability{"completion", "tools"}
}

// Models holds the configuration for all bases and models.
type Models struct {
	Bases     []BaseModel    `koanf:"bases"`
	Models    []ModelInfo    `koanf:"models"`
	mapBases  map[string]int `koanf:"-"`
	mapModels map[string]int `koanf:"-"`
}

func DefaultModelInfo() *ModelInfo {
	return &ModelInfo{
		Name: "default",
		BaseModelConfig: BaseModelConfig{
			Capabilities: []model.Capability{"completion", "tools"},
			InputTokens:  0,
			OutputTokens: 0,
		},
		Base:      nil,
		baseModel: nil,
	}
}

func (m *Models) GetModel(name string) (*ModelInfo, error) {
	if idx, exists := m.mapModels[name]; exists {
		return &m.Models[idx], nil
	}
	return nil, fmt.Errorf("model '%s' not found", name)
}

func LoadModels(path string) (*Models, error) {
	var models Models
	k := koanf.New("_")

	// Load models.yml file
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading models.yml: %w", err)
	}

	// Unmarshal the configuration into the Models struct
	if err := k.Unmarshal("", &models); err != nil {
		return nil, fmt.Errorf("error unmarshalling models config: %w", err)
	}

	models.mapBases = make(map[string]int)
	for i, base := range models.Bases {
		models.mapBases[base.Name] = i
	}

	models.mapModels = make(map[string]int)
	for i := range models.Models {
		model := &models.Models[i]
		models.mapModels[model.Name] = i
		if model.Base != nil {
			if baseIdx, exists := models.mapBases[*model.Base]; exists {
				model.baseModel = &models.Bases[baseIdx]
			} else {
				slog.Warn("Model references non-existent base", "model", model.Name, "base", *model.Base)
			}
		}
	}

	return &models, nil
}
