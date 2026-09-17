package provider

import (
	"fmt"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

type Builder func(Settings) (Provider, error)

type Factory struct {
	builders map[string]Builder
}

func NewFactory() *Factory {
	return &Factory{builders: map[string]Builder{
		"mock":   newMockProvider,
		"gemini": newGeminiProvider,
	}}
}

func (f *Factory) Create(settings Settings) (Provider, error) {
	builder, exists := f.builders[settings.Name]
	if !exists {
		return nil, chat.NewError("invalid_provider", "AI_SCENERY_GO_PROVIDER 必须是 mock 或 gemini", 500)
	}
	provider, err := builder(settings)
	if err != nil {
		return nil, fmt.Errorf("创建 %s provider 失败: %w", settings.Name, err)
	}
	return provider, nil
}
