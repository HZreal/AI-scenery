package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/HZreal/AI-scenery/demos/01-llm-api/src/gin_api/internal/provider"
)

type Config struct {
	Address       string
	MaxInputChars int
	Provider      provider.Settings
}

func Load() (Config, error) {
	loadEnvFile(".env.local")
	limit, err := positiveInt("MAX_INPUT_CHARS", 12000)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Address:       env("GIN_LLM_API_ADDRESS", "127.0.0.1:8002"),
		MaxInputChars: limit,
		Provider: provider.Settings{
			Name:          strings.ToLower(env("AI_SCENERY_GO_PROVIDER", "mock")),
			GeminiAPIKey:  strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
			GeminiModel:   env("GEMINI_MODEL", "gemini-3.8-flash"),
			OpenAIAPIKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
			OpenAIModel:   env("OPENAI_MODEL", "gpt-5.5"),
			OpenAIBaseURL: strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")),
		},
	}, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func positiveInt(key string, fallback int) (int, error) {
	value := env(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s 必须是正整数", key)
	}
	return parsed, nil
}

// loadEnvFile 仅补充未设置的变量，命令行或 IDE 显式配置优先。
func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found || os.Getenv(strings.TrimSpace(key)) != "" {
			continue
		}
		_ = os.Setenv(strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), "\"'"))
	}
}
