package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func loadConfig() {
	// Check for custom config path via environment variable first
	if configPath := os.Getenv("NAS_CONFIG"); configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			loadEnvFile(configPath)
			return
		}
	}
	
	configPaths := []string{
		".nasrc",                                      // 1. Working directory
		filepath.Join(os.Getenv("HOME"), ".nasrc"), // 2. Home directory
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			loadEnvFile(path)
			break
		}
	}
}

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
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}