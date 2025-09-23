package yandexapi

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// GetIAMTokenFromEnv читает токен из файла или переменной окружения
func GetIAMTokenFromEnv() (string, error) {
	// 1) Сначала проверяем переменную окружения YC_IAM_TOKEN
	if tok := strings.TrimSpace(os.Getenv("YC_IAM_TOKEN")); tok != "" {
		return tok, nil
	}

	// 2) Потом проверяем файл
	if path := strings.TrimSpace(os.Getenv("YC_IAM_TOKEN_FILE")); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read token file: %w", err)
		}
		tok := strings.TrimSpace(string(b))
		if tok == "" {
			return "", errors.New("token file is empty")
		}
		return tok, nil
	}

	return "", errors.New("no IAM token provided (set YC_IAM_TOKEN or YC_IAM_TOKEN_FILE)")
}
