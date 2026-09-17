package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HZreal/AI-scenery/demos/02-prompt-context/src/gin_prompt_context/internal/prompt"
)

type server struct {
	generator       prompt.Generator
	maxContextChars int
}

type compareInput struct {
	Task           string   `json:"task"`
	Context        string   `json:"context"`
	PromptVersions []string `json:"prompt_versions"`
}

func NewServer(generator prompt.Generator, maxContextChars int) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := server{generator: generator, maxContextChars: maxContextChars}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/openapi.json", func(c *gin.Context) { c.JSON(http.StatusOK, openAPISpec) })
	router.GET("/docs", swaggerUI)
	router.POST("/api/prompt-context", server.compare)
	return router
}

func (s server) compare(c *gin.Context) {
	traceID := newTraceID()
	var input compareInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, promptError("invalid_request", "请求体必须是 JSON 对象", 400), traceID)
		return
	}
	started := time.Now()
	result, model, err := prompt.Compare(c.Request.Context(), s.generator, prompt.Input{Task: input.Task, Context: input.Context, Versions: input.PromptVersions}, s.maxContextChars)
	if err != nil {
		respondError(c, err, traceID)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result": result,
		"metadata": gin.H{
			"model":             model,
			"latency_ms":        time.Since(started).Milliseconds(),
			"prompt_versions":   len(result.Comparisons),
			"cost_estimate_usd": nil,
			"trace_id":          traceID,
		},
	})
}

func respondError(c *gin.Context, err error, traceID string) {
	appError, ok := err.(*prompt.AppError)
	if !ok {
		appError = promptError("internal_error", "服务内部错误", 500)
	}
	c.JSON(appError.StatusCode, gin.H{"error": appError.Message, "code": appError.Code, "metadata": gin.H{"trace_id": traceID}})
}

func promptError(code, message string, statusCode int) *prompt.AppError {
	return &prompt.AppError{Code: code, Message: message, StatusCode: statusCode}
}

func newTraceID() string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "trace-unavailable"
	}
	return "trace-" + hex.EncodeToString(bytes)
}
