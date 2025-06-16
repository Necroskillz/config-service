package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteFallbackFile(config *Config, fileName string, data any) error {
	if !config.IsFallbackFileEnabled() {
		return fmt.Errorf("WriteFallbackFile called but fallback file is not enabled")
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	filePath := filepath.Join(config.FallbackFileLocation, fileName)

	if err := os.MkdirAll(config.FallbackFileLocation, 0755); err != nil {
		return fmt.Errorf("failed to create fallback file location: %w", err)
	}

	return os.WriteFile(filePath, bytes, 0644)
}

func ReadFallbackFile[T any](config *Config, fileName string) (*T, error) {
	if !config.IsFallbackFileEnabled() {
		return nil, fmt.Errorf("ReadFallbackFile called but fallback file is not enabled")
	}

	filePath := filepath.Join(config.FallbackFileLocation, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("fallback file does not exist: %w", err)
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fallback file: %w", err)
	}

	var data T

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &data, nil
}
