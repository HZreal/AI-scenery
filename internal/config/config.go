package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/HZreal/AI-scenery/internal/llm"
)

type Config struct {
	Address         string
	MaxInputChars   int
	MaxContextChars int
	Provider        llm.Settings
}

func Load() (Config, error) {
	loadEnvFile(".env.local")
	maxInputChars, err := positiveInt("MAX_INPUT_CHARS", 12000)
	if err != nil {
		return Config{}, err
	}
	maxContextChars, err := positiveInt("PROMPT_CONTEXT_MAX_CHARS", 6000)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Address:         env("AI_SCENERY_ADDRESS", "127.0.0.1:8002"),
		MaxInputChars:   maxInputChars,
		MaxContextChars: maxContextChars,
		Provider: llm.Settings{
			Name:          strings.ToLower(env("AI_SCENERY_PROVIDER", "mock")),
			GeminiAPIKey:  strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
			GeminiModel:   env("GEMINI_MODEL", "gemini-2.5-flash-lite"),
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
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s 必须是正整数", key)
	}
	return value, nil
}

// loadEnvFile only fills variables not explicitly set by the shell or IDE.
func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, found := strings.Cut(strings.TrimSpace(scanner.Text()), "=")
		key = strings.TrimSpace(key)
		if !found || key == "" || strings.HasPrefix(key, "#") || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, strings.Trim(strings.TrimSpace(value), "\"'"))
	}
}
