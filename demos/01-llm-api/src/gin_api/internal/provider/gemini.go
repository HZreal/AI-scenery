package provider

import (
	"context"
	"encoding/json"
	"strings"

	"google.golang.org/genai"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

type geminiProvider struct {
	client *genai.Client
	model  string
}

func newGeminiProvider(settings Settings) (Provider, error) {
	if strings.TrimSpace(settings.APIKey) == "" {
		return nil, chat.NewError("missing_api_key", "AI_SCENERY_GO_PROVIDER=gemini 时必须设置 GEMINI_API_KEY", 500)
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  settings.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, chat.NewError("provider_initialization_failed", "Gemini 客户端初始化失败: "+err.Error(), 502)
	}
	return &geminiProvider{client: client, model: settings.Model}, nil
}

func (p *geminiProvider) Generate(ctx context.Context, request chat.Request) (chat.Result, error) {
	contents, options := geminiInput(request)
	response, err := p.client.Models.GenerateContent(ctx, p.model, contents, options)
	if err != nil {
		return chat.Result{}, chat.NewError("provider_api_error", "Gemini API 调用失败: "+err.Error(), 502)
	}
	return p.resultFrom(response, request.Mode)
}

func (p *geminiProvider) Stream(ctx context.Context, request chat.Request, emit func(string) error) (chat.Result, error) {
	contents, options := geminiInput(chat.Request{Messages: request.Messages, Mode: chat.ModeText})
	var last *genai.GenerateContentResponse
	for response, err := range p.client.Models.GenerateContentStream(ctx, p.model, contents, options) {
		if err != nil {
			return chat.Result{}, chat.NewError("provider_stream_error", "Gemini 流式调用失败: "+err.Error(), 502)
		}
		last = response
		if text := response.Text(); text != "" {
			if err := emit(text); err != nil {
				return chat.Result{}, err
			}
		}
	}
	if last == nil {
		return chat.Result{}, chat.NewError("provider_response_invalid", "Gemini 未返回流式内容", 502)
	}
	return p.resultFrom(last, chat.ModeText)
}

func (p *geminiProvider) resultFrom(response *genai.GenerateContentResponse, mode chat.Mode) (chat.Result, error) {
	text := response.Text()
	if text == "" {
		return chat.Result{}, chat.NewError("provider_response_invalid", "Gemini 响应中没有文本内容", 502)
	}
	result := chat.Result{Value: text, Model: p.model}
	if response.UsageMetadata != nil {
		result.InputTokens = &response.UsageMetadata.PromptTokenCount
		result.OutputTokens = &response.UsageMetadata.CandidatesTokenCount
	}
	if mode == chat.ModeJSON {
		var value map[string]any
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			return chat.Result{}, chat.NewError("provider_json_invalid", "Gemini 未返回有效 JSON 对象", 502)
		}
		result.Value = value
	}
	return result, nil
}

func geminiInput(request chat.Request) ([]*genai.Content, *genai.GenerateContentConfig) {
	contents := make([]*genai.Content, 0, len(request.Messages))
	systemParts := make([]string, 0)
	for _, message := range request.Messages {
		if message.Role == "system" {
			systemParts = append(systemParts, message.Content)
			continue
		}
		var role genai.Role = genai.RoleUser
		if message.Role == "assistant" {
			role = genai.RoleModel
		}
		contents = append(contents, genai.NewContentFromText(message.Content, role))
	}
	options := &genai.GenerateContentConfig{}
	if len(systemParts) > 0 {
		options.SystemInstruction = genai.NewContentFromText(strings.Join(systemParts, "\n"), genai.RoleUser)
	}
	if request.Mode == chat.ModeJSON {
		options.ResponseMIMEType = "application/json"
	}
	return contents, options
}
