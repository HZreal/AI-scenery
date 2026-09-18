package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	llmapi "github.com/HZreal/AI-scenery/demos/01-llm-api/lesson"
	promptcontext "github.com/HZreal/AI-scenery/demos/02-prompt-context/lesson"
	"github.com/HZreal/AI-scenery/internal/catalog"
	"github.com/HZreal/AI-scenery/internal/config"
	"github.com/HZreal/AI-scenery/internal/llm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	provider, err := llm.NewFactory().Create(cfg.Provider)
	if err != nil {
		log.Fatal(err)
	}
	router := newRouter(provider, cfg)
	log.Printf("Serving AI Scenery on http://%s", cfg.Address)
	if err := router.Run(cfg.Address); err != nil {
		log.Fatal(err)
	}
}

func newRouter(provider llm.Provider, cfg config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/web/") {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	})
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/web/") })
	router.StaticFS("/web", http.Dir("web"))
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/api/catalog", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"demos": catalog.All()}) })
	llmapi.Register(router.Group("/api/llm"), provider, cfg.MaxInputChars)
	promptcontext.Register(router.Group("/api/prompt-context"), provider, cfg.MaxContextChars)
	router.GET("/openapi.json", func(c *gin.Context) { c.JSON(http.StatusOK, mergedOpenAPI()) })
	router.GET("/docs", swaggerUI)
	return router
}

func mergedOpenAPI() gin.H {
	llmSpec := llmapi.OpenAPISpec
	promptSpec := promptcontext.OpenAPISpec
	paths := gin.H{
		"/health": gin.H{
			"get": gin.H{
				"summary":   "检查服务状态",
				"responses": gin.H{"200": gin.H{"description": "服务正常"}},
			},
		},
	}
	for path, operation := range llmSpec["paths"].(gin.H) {
		if path != "/health" {
			paths[path] = operation
		}
	}
	for path, operation := range promptSpec["paths"].(gin.H) {
		if path != "/health" {
			paths[path] = operation
		}
	}
	return gin.H{
		"openapi": "3.0.3",
		"info":    gin.H{"title": "AI Scenery Learning API", "version": "1.0.0", "description": "AI 与 AI Agent 后端学习 Demo。"},
		"paths":   paths,
		"components": gin.H{"schemas": gin.H{
			"ChatRequest":          llmSpec["components"].(gin.H)["schemas"].(gin.H)["ChatRequest"],
			"Message":              llmSpec["components"].(gin.H)["schemas"].(gin.H)["Message"],
			"PromptCompareRequest": promptSpec["components"].(gin.H)["schemas"].(gin.H)["PromptCompareRequest"],
		}},
	}
}

func swaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>AI Scenery API 文档</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:"/openapi.json",dom_id:"#swagger-ui"})</script></body></html>`))
}
