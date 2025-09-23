package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/studio-b12/gowebdav"
)

// WebDAV configuration
type WebDAVConfig struct {
	URL      string
	Username string
	Password string
	BasePath string
}

// WebDAVClient wraps the gowebdav client with our specific functionality
type WebDAVClient struct {
	client *gowebdav.Client
	config *WebDAVConfig
}

// NewWebDAVClient creates a new WebDAV client
func NewWebDAVClient(config *WebDAVConfig) *WebDAVClient {
	client := gowebdav.NewClient(config.URL, config.Username, config.Password)
	
	// Set up TLS config similar to existing client
	httpClient := GetClient()
	client.SetTransport(httpClient.Transport)
	
	return &WebDAVClient{
		client: client,
		config: config,
	}
}

// GetLastSyncDate retrieves the last sync date from WebDAV server
func (w *WebDAVClient) GetLastSyncDate() (time.Time, error) {
	syncFilePath := path.Join(w.config.BasePath, "last_sync.txt")
	
	data, err := w.client.Read(syncFilePath)
	if err != nil {
		// If file doesn't exist, return epoch time (sync everything)
		log.Printf("Last sync file not found, syncing all highlights: %v", err)
		return time.Unix(0, 0), nil
	}
	
	dateStr := strings.TrimSpace(string(data))
	return time.Parse(time.RFC3339, dateStr)
}

// UpdateLastSyncDate updates the last sync date on WebDAV server
func (w *WebDAVClient) UpdateLastSyncDate(syncTime time.Time) error {
	syncFilePath := path.Join(w.config.BasePath, "last_sync.txt")
	dateStr := syncTime.Format(time.RFC3339)
	
	return w.client.Write(syncFilePath, []byte(dateStr), 0644)
}

// GetBookFile downloads a book's markdown file from WebDAV server
func (w *WebDAVClient) GetBookFile(bookTitle string) ([]byte, error) {
	filename := sanitizeFilename(bookTitle) + ".md"
	filePath := path.Join(w.config.BasePath, filename)
	
	data, err := w.client.Read(filePath)
	if err != nil {
		// File doesn't exist, return empty content
		return []byte{}, nil
	}
	
	return data, nil
}

// SaveBookFile atomically saves a book's markdown file to WebDAV server
func (w *WebDAVClient) SaveBookFile(bookTitle string, content []byte) error {
	filename := sanitizeFilename(bookTitle) + ".md"
	filePath := path.Join(w.config.BasePath, filename)
	backupPath := filePath + ".backup"
	tempPath := filePath + ".tmp"
	
	// Step 1: Create backup if file exists
	_, err := w.client.Stat(filePath)
	if err == nil {
		existingData, err := w.client.Read(filePath)
		if err == nil {
			if err := w.client.Write(backupPath, existingData, 0644); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}
	}
	
	// Step 2: Write to temporary file
	if err := w.client.Write(tempPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Step 3: Move temp file to final location (atomic operation)
	if err := w.client.Rename(tempPath, filePath, false); err != nil {
		// Clean up temp file on failure
		w.client.Remove(tempPath)
		return fmt.Errorf("failed to move temp file to final location: %w", err)
	}
	
	// Step 4: Remove backup on success
	w.client.Remove(backupPath)
	
	return nil
}

// sanitizeFilename removes or replaces characters that are problematic in filenames
func sanitizeFilename(filename string) string {
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	
	sanitized := replacer.Replace(filename)
	
	// Trim spaces and dots from the ends
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure filename is not empty
	if sanitized == "" {
		sanitized = "unknown"
	}
	
	return sanitized
}