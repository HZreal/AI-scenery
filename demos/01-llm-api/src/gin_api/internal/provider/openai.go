package provider

import (
	"context"
	"encoding/json"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

type openAIProvider struct {
	client openai.Client
	model  string
}

func newOpenAIProvider(settings Settings) (Provider, error) {
	if strings.TrimSpace(settings.OpenAIAPIKey) == "" {
		return nil, chat.NewError("missing_api_key", "AI_SCENERY_GO_PROVIDER=openai 时必须设置 OPENAI_API_KEY", 500)
	}
	options := []option.RequestOption{option.WithAPIKey(settings.OpenAIAPIKey)}
	if settings.OpenAIBaseURL != "" {
		options = append(options, option.WithBaseURL(settings.OpenAIBaseURL))
	}
	return &openAIProvider{client: openai.NewClient(options...), model: settings.OpenAIModel}, nil
}

func (p *openAIProvider) Generate(ctx context.Context, request chat.Request) (chat.Result, error) {
	response, err := p.client.Responses.New(ctx, openAIParams(request, p.model))
	if err != nil {
		return chat.Result{}, chat.NewError("provider_api_error", "OpenAI API 调用失败: "+err.Error(), 502)
	}
	return p.resultFrom(response, request.Mode)
}

func (p *openAIProvider) Stream(ctx context.Context, request chat.Request, emit func(string) error) (chat.Result, error) {
	stream := p.client.Responses.NewStreaming(ctx, openAIParams(chat.Request{Messages: request.Messages, Mode: chat.ModeText}, p.model))
	defer stream.Close()

	var completed *responses.Response
	for stream.Next() {
		event := stream.Current()
		switch event.Type {
		case "response.output_text.delta":
			if err := emit(event.AsResponseOutputTextDelta().Delta); err != nil {
				return chat.Result{}, err
			}
		case "response.completed":
			response := event.AsResponseCompleted().Response
			completed = &response
		}
	}
	if err := stream.Err(); err != nil {
		return chat.Result{}, chat.NewError("provider_stream_error", "OpenAI 流式调用失败: "+err.Error(), 502)
	}
	if completed == nil {
		return chat.Result{}, chat.NewError("provider_response_invalid", "OpenAI 未返回完成事件", 502)
	}
	return p.resultFrom(completed, chat.ModeText)
}

func (p *openAIProvider) resultFrom(response *responses.Response, mode chat.Mode) (chat.Result, error) {
	text := response.OutputText()
	if text == "" {
		return chat.Result{}, chat.NewError("provider_response_invalid", "OpenAI 响应中没有文本内容", 502)
	}
	inputTokens := int32(response.Usage.InputTokens)
	outputTokens := int32(response.Usage.OutputTokens)
	result := chat.Result{Value: text, Model: p.model, InputTokens: &inputTokens, OutputTokens: &outputTokens}
	if mode == chat.ModeJSON {
		var value map[string]any
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			return chat.Result{}, chat.NewError("provider_json_invalid", "OpenAI 未返回有效 JSON 对象", 502)
		}
		result.Value = value
	}
	return result, nil
}

func openAIParams(request chat.Request, model string) responses.ResponseNewParams {
	items := make(responses.ResponseInputParam, 0, len(request.Messages))
	for _, message := range request.Messages {
		role := responses.EasyInputMessageRole(message.Role)
		items = append(items, responses.ResponseInputItemUnionParam{OfMessage: &responses.EasyInputMessageParam{
			Role:    role,
			Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(message.Content)},
		}})
	}
	params := responses.ResponseNewParams{
		Model: shared.ResponsesModel(model),
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: items},
	}
	if request.Mode == chat.ModeJSON {
		params.Text = responses.ResponseTextConfigParam{Format: responses.ResponseFormatTextConfigUnionParam{OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
			Name:   "llm_api_demo",
			Strict: openai.Bool(true),
			Schema: structuredOutputSchema(),
		}}}
	}
	return params
}

func structuredOutputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"topic":      map[string]string{"type": "string"},
			"summary":    map[string]string{"type": "string"},
			"key_points": map[string]any{"type": "array", "items": map[string]string{"type": "string"}},
		},
		"required": []string{"topic", "summary", "key_points"},
	}
}
