package prompt

import (
	"context"
	"strings"
	"testing"
)

type fakeGenerator struct{}

func (fakeGenerator) Generate(_ context.Context, system, user string) (Generation, error) {
	return Generation{Text: "测试输出", Model: "fake-gemini"}, nil
}

func TestGeminiRetryOptions(t *testing.T) {
	opts := geminiRetryOptions()
	if opts.Attempts == nil || *opts.Attempts != 3 {
		t.Fatalf("unexpected retry attempts: %#v", opts.Attempts)
	}
	if len(opts.HTTPStatusCodes) != 2 || opts.HTTPStatusCodes[0] != 429 || opts.HTTPStatusCodes[1] != 503 {
		t.Fatalf("unexpected retryable status codes: %#v", opts.HTTPStatusCodes)
	}
}

func TestCompareBuildsVersionsAndTrimsContext(t *testing.T) {
	result, model, err := Compare(context.Background(), fakeGenerator{}, Input{
		Task:    "解释上下文裁剪",
		Context: strings.Repeat("上下文", 20),
	}, 24)
	if err != nil {
		t.Fatal(err)
	}
	if model != "fake-gemini" || len(result.Comparisons) != 2 || !result.Context.Truncated {
		t.Fatalf("unexpected comparison result: %#v", result)
	}
	if !strings.Contains(result.Comparisons[1].UserPrompt, "<context>") || !strings.Contains(result.Comparisons[1].UserPrompt, "<task>") {
		t.Fatalf("grounded prompt is missing boundaries: %s", result.Comparisons[1].UserPrompt)
	}
}

func TestCompareRejectsUnknownPromptVersion(t *testing.T) {
	_, _, err := Compare(context.Background(), fakeGenerator{}, Input{Task: "测试", Versions: []string{"unknown"}}, 100)
	if err == nil {
		t.Fatal("expected invalid prompt version error")
	}
}
