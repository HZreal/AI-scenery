package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
)

type testProvider struct{}

func (testProvider) Generate(_ context.Context, _ chat.Request) (chat.Result, error) {
	return chat.Result{Value: "测试响应", Model: "test-model"}, nil
}

func (testProvider) Stream(_ context.Context, _ chat.Request, emit func(string) error) (chat.Result, error) {
	if err := emit("测试 "); err != nil {
		return chat.Result{}, err
	}
	return chat.Result{Value: "测试", Model: "test-model"}, nil
}

func TestChatReturnsStableResponseShape(t *testing.T) {
	router := NewServer(testProvider{}, 100)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"message":"什么是 Agent？"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"result":"测试响应"`) || !strings.Contains(recorder.Body.String(), `"trace_id"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestChatRejectsEmptyMessage(t *testing.T) {
	router := NewServer(testProvider{}, 100)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"message":"  "}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"empty_message"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestOpenAPISpecIsAvailable(t *testing.T) {
	router := NewServer(testProvider{}, 100)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"openapi":"3.0.3"`) || !strings.Contains(recorder.Body.String(), `"/api/chat"`) {
		t.Fatalf("unexpected OpenAPI response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
