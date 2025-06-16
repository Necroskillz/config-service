package internal

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type ConfigurationFallbackFile struct {
	config *Config
}

func NewConfigurationFallbackFile(config *Config) *ConfigurationFallbackFile {
	return &ConfigurationFallbackFile{
		config: config,
	}
}

func (c *ConfigurationFallbackFile) getSnapshotFileName() string {
	fileName := strings.Builder{}
	fileName.WriteString("configuration")

	for _, property := range slices.Sorted(maps.Keys(c.config.StaticVariation)) {
		value := c.config.StaticVariation[property]
		fileName.WriteString(fmt.Sprintf("_%s-%s", property, value))
	}

	fileName.WriteString(".json")

	return fileName.String()
}

func (c *ConfigurationFallbackFile) Write(snapshot *ConfigurationSnapshot) error {
	if err := WriteFallbackFile(c.config, c.getSnapshotFileName(), snapshot); err != nil {
		return fmt.Errorf("failed to write configuration fallback file: %w", err)
	}

	return nil
}

func (c *ConfigurationFallbackFile) Read() (*ConfigurationSnapshot, error) {
	snapshot, err := ReadFallbackFile[ConfigurationSnapshot](c.config, c.getSnapshotFileName())
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration fallback file: %w", err)
	}

	return snapshot, nil
}
