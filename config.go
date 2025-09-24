package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Configuration paths
const (
	configPath         = "/mnt/onboard/.adds/go-kobo-sync/config"
	templatePath       = "/mnt/onboard/.adds/go-kobo-sync/template.md"
	headerTemplatePath = "/mnt/onboard/.adds/go-kobo-sync/header_template.md"
	dbLocation         = "/mnt/onboard/.kobo/KoboReader.sqlite"
)

// Config represents the application configuration
type Config struct {
	WebDAV         *WebDAVConfig
	Template       string
	HeaderTemplate string
}

// LoadConfig loads configuration from the config file
func LoadConfig() (*Config, error) {
	config := &Config{
		WebDAV: &WebDAVConfig{},
	}

	// Load WebDAV configuration
	if err := loadWebDAVConfig(config.WebDAV); err != nil {
		return nil, fmt.Errorf("failed to load WebDAV config: %w", err)
	}

	// Load template if exists
	if templateData, err := os.ReadFile(templatePath); err == nil {
		config.Template = string(templateData)
	}
	// Load header template if exists
	if templateData, err := os.ReadFile(headerTemplatePath); err == nil {
		config.HeaderTemplate = string(templateData)
	}

	return config, nil
}

// loadWebDAVConfig loads WebDAV configuration from file
func loadWebDAVConfig(webdavConfig *WebDAVConfig) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("unable to open config file %s: %w", configPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "webdav_url":
			webdavConfig.URL = value
		case "webdav_username":
			webdavConfig.Username = value
		case "webdav_password":
			webdavConfig.Password = value
		case "webdav_path":
			webdavConfig.BasePath = value
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Validate required fields
	if webdavConfig.URL == "" {
		return fmt.Errorf("webdav_url is required")
	}
	if webdavConfig.Username == "" {
		return fmt.Errorf("webdav_username is required")
	}
	if webdavConfig.Password == "" {
		return fmt.Errorf("webdav_password is required")
	}
	if webdavConfig.BasePath == "" {
		webdavConfig.BasePath = "/kobo-highlights"
	}

	return nil
}
