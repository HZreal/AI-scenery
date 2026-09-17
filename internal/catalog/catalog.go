package catalog

type Demo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Streaming   bool   `json:"streaming"`
	RequestPath string `json:"request_example"`
	Document    string `json:"document"`
}

func All() []Demo {
	return []Demo{
		{ID: "llm-api", Name: "01. LLM 应用基础", Endpoint: "/api/llm/chat", Method: "POST", Streaming: true, RequestPath: "/demos/01-llm-api/examples/request.json", Document: "/demos/01-llm-api/README.md"},
		{ID: "prompt-context", Name: "02. Prompt 与上下文工程", Endpoint: "/api/prompt-context", Method: "POST", Streaming: false, RequestPath: "/demos/02-prompt-context/examples/request.json", Document: "/demos/02-prompt-context/README.md"},
	}
}
