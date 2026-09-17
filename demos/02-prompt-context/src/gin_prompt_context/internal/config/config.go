package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address         string
	MaxContextChars int
	GeminiAPIKey    string
	GeminiModel     string
}

func Load() (Config, error) {
	loadEnvFile(".env.local")
	maxContextChars, err := positiveInt("PROMPT_CONTEXT_MAX_CHARS", 6000)
	if err != nil {
		return Config{}, err
	}
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return Config{}, fmt.Errorf("GEMINI_API_KEY 未设置")
	}
	return Config{
		Address:         env("PROMPT_CONTEXT_ADDRESS", "127.0.0.1:8003"),
		MaxContextChars: maxContextChars,
		GeminiAPIKey:    apiKey,
		GeminiModel:     env("GEMINI_MODEL", "gemini-2.5-flash-lite"),
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

// loadEnvFile 只补充缺失变量，IDE 或命令行中显式设置的值优先。
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
		key = strings.TrimSpace(key)
		if !found || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, strings.Trim(strings.TrimSpace(value), "\"'"))
	}
}
