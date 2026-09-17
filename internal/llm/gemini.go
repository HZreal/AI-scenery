package llm

import (
	"context"
	"encoding/json"
	"strings"

	"google.golang.org/genai"
)

type geminiProvider struct {
	client *genai.Client
	model  string
}

func newGeminiProvider(settings Settings) (Provider, error) {
	if strings.TrimSpace(settings.GeminiAPIKey) == "" {
		return nil, NewError("missing_api_key", "AI_SCENERY_PROVIDER=gemini 时必须设置 GEMINI_API_KEY", 500)
	}
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  settings.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			// Do not retry invalid requests; only absorb short provider overloads.
			RetryOptions: &genai.HTTPRetryOptions{
				Attempts:        genai.Ptr(int32(3)),
				InitialDelay:    genai.Ptr(0.5),
				MaxDelay:        genai.Ptr(2.0),
				HTTPStatusCodes: []int32{429, 503},
			},
		},
	})
	if err != nil {
		return nil, NewError("provider_initialization_failed", "Gemini 客户端初始化失败: "+err.Error(), 502)
	}
	return &geminiProvider{client: client, model: settings.GeminiModel}, nil
}

func (p *geminiProvider) Generate(ctx context.Context, request Request) (Result, error) {
	contents, options := geminiInput(request)
	response, err := p.client.Models.GenerateContent(ctx, p.model, contents, options)
	if err != nil {
		return Result{}, NewError("provider_api_error", "Gemini API 调用失败: "+err.Error(), 502)
	}
	return p.resultFrom(response, request.Mode)
}

func (p *geminiProvider) Stream(ctx context.Context, request Request, emit func(string) error) (Result, error) {
	contents, options := geminiInput(Request{Messages: request.Messages, Mode: ModeText})
	var last *genai.GenerateContentResponse
	for response, err := range p.client.Models.GenerateContentStream(ctx, p.model, contents, options) {
		if err != nil {
			return Result{}, NewError("provider_stream_error", "Gemini 流式调用失败: "+err.Error(), 502)
		}
		last = response
		if text := response.Text(); text != "" {
			if err := emit(text); err != nil {
				return Result{}, err
			}
		}
	}
	if last == nil {
		return Result{}, NewError("provider_response_invalid", "Gemini 未返回流式内容", 502)
	}
	return p.resultFrom(last, ModeText)
}

func (p *geminiProvider) resultFrom(response *genai.GenerateContentResponse, mode Mode) (Result, error) {
	text := response.Text()
	if text == "" {
		return Result{}, NewError("provider_response_invalid", "Gemini 响应中没有文本内容", 502)
	}
	result := Result{Value: text, Model: p.model}
	if response.UsageMetadata != nil {
		result.InputTokens = &response.UsageMetadata.PromptTokenCount
		result.OutputTokens = &response.UsageMetadata.CandidatesTokenCount
	}
	if mode == ModeJSON {
		var value map[string]any
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			return Result{}, NewError("provider_json_invalid", "Gemini 未返回有效 JSON 对象", 502)
		}
		result.Value = value
	}
	return result, nil
}

func geminiInput(request Request) ([]*genai.Content, *genai.GenerateContentConfig) {
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
	if request.Mode == ModeJSON {
		// MIME 类型与 Schema 要同时提供，避免模型返回 Markdown 或任意 JSON 形状。
		options.ResponseMIMEType = "application/json"
		options.ResponseSchema = geminiStructuredOutputSchema()
	}
	return contents, options
}

func geminiStructuredOutputSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"topic":      {Type: genai.TypeString},
			"summary":    {Type: genai.TypeString},
			"key_points": {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
		},
		Required: []string{"topic", "summary", "key_points"},
	}
}
