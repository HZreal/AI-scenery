package api

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/chat"
	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/provider"
)

type server struct {
	provider      provider.Provider
	maxInputChars int
}

type chatInput struct {
	Messages []chat.Message `json:"messages"`
	Message  string         `json:"message"`
	Mode     string         `json:"mode"`
	Stream   bool           `json:"stream"`
}

func NewServer(llm provider.Provider, maxInputChars int) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := server{provider: llm, maxInputChars: maxInputChars}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/openapi.json", func(c *gin.Context) { c.JSON(http.StatusOK, openAPISpec) })
	router.GET("/docs", swaggerUI)
	router.POST("/api/chat", server.chat)
	return router
}

func (s server) chat(c *gin.Context) {
	traceID := newTraceID()
	var input chatInput
	if err := c.ShouldBindJSON(&input); err != nil {
		s.respondError(c, chat.NewError("invalid_request", "请求体必须是 JSON 对象", 400), traceID)
		return
	}
	request, err := chat.NewRequest(input.Messages, input.Message, input.Mode)
	if err != nil {
		s.respondError(c, err, traceID)
		return
	}
	if request.InputCharacters() > s.maxInputChars {
		s.respondError(c, chat.NewError("context_limit_exceeded", "输入超过 MAX_INPUT_CHARS 限制", 400), traceID)
		return
	}
	if input.Stream {
		s.stream(c, request, traceID)
		return
	}

	started := time.Now()
	result, err := s.provider.Generate(c.Request.Context(), request)
	if err != nil {
		s.respondError(c, err, traceID)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result.Value, "metadata": metadata(result, request.InputCharacters(), traceID, time.Since(started))})
}

func (s server) stream(c *gin.Context, request chat.Request, traceID string) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Stream(func(writer io.Writer) bool {
		c.SSEvent("metadata", gin.H{"trace_id": traceID})
		started := time.Now()
		result, err := s.provider.Stream(c.Request.Context(), request, func(delta string) error {
			c.SSEvent("delta", gin.H{"delta": delta})
			c.Writer.Flush()
			return nil
		})
		if err != nil {
			c.SSEvent("error", errorBody(err, traceID))
			return false
		}
		c.SSEvent("completed", gin.H{"metadata": metadata(result, request.InputCharacters(), traceID, time.Since(started))})
		return false
	})
}

func (s server) respondError(c *gin.Context, err error, traceID string) {
	apiError := asChatError(err)
	c.JSON(apiError.StatusCode, errorBody(apiError, traceID))
}

func errorBody(err error, traceID string) gin.H {
	apiError := asChatError(err)
	return gin.H{"error": apiError.Message, "code": apiError.Code, "metadata": gin.H{"trace_id": traceID}}
}

func asChatError(err error) *chat.Error {
	if typed, ok := err.(*chat.Error); ok {
		return typed
	}
	return chat.NewError("internal_error", "服务内部错误", 500)
}

func metadata(result chat.Result, inputCharacters int, traceID string, latency time.Duration) gin.H {
	return gin.H{
		"model":            result.Model,
		"latency_ms":       latency.Milliseconds(),
		"input_characters": inputCharacters,
		"usage": gin.H{
			"input_tokens":  result.InputTokens,
			"output_tokens": result.OutputTokens,
		},
		"cost_estimate_usd": nil,
		"trace_id":          traceID,
	}
}

func newTraceID() string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "trace-unavailable"
	}
	return "trace-" + hex.EncodeToString(bytes)
}
