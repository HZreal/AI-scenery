package llm

import (
	"fmt"
)

type Builder func(Settings) (Provider, error)

type Factory struct {
	builders map[string]Builder
}

func NewFactory() *Factory {
	return &Factory{builders: map[string]Builder{
		"mock":   newMockProvider,
		"gemini": newGeminiProvider,
		"openai": newOpenAIProvider,
	}}
}

func (f *Factory) Create(settings Settings) (Provider, error) {
	builder, exists := f.builders[settings.Name]
	if !exists {
		return nil, NewError("invalid_provider", "AI_SCENERY_PROVIDER 必须是 mock、gemini 或 openai", 500)
	}
	provider, err := builder(settings)
	if err != nil {
		return nil, fmt.Errorf("创建 %s provider 失败: %w", settings.Name, err)
	}
	return provider, nil
}
