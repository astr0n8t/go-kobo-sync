package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/studio-b12/gowebdav"
)

const lastSyncFile = ".sync_status"

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
	syncFilePath := path.Join(w.config.BasePath, lastSyncFile)

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
	syncFilePath := path.Join(w.config.BasePath, lastSyncFile)
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
func (w *WebDAVClient) SaveBookFile(bookTitle string, header []byte, content []byte) error {
	filename := sanitizeFilename(bookTitle) + ".md"
	filePath := path.Join(w.config.BasePath, filename)
	backupPath := filePath + ".backup"
	tempPath := filePath + ".tmp"
	var existingData []byte

	// Step 1: Create backup if file exists
	_, err := w.client.Stat(filePath)
	if err == nil {
		existingData, err = w.client.Read(filePath)
		if err != nil {
			return fmt.Errorf("failed to get previous data : %w", err)
		} else {
			header = existingData
		}
	}

	content = append(header, content...)

	// Step 2: Write to temporary file
	if err := w.client.Write(tempPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Step 3: Move temp file to final location (atomic operation)
	if err := w.client.Rename(tempPath, filePath, true); err != nil {
		// Clean up temp file on failure
		w.client.Remove(tempPath)
		return fmt.Errorf("failed to move temp file to final location: %w", err)
	}

	// Step 4: Remove backup on success
	w.client.Remove(backupPath)

	return nil
}

// sanitizeFilename removes or replaces characters that are problematic in filenames
// Limits to 128 characters, only ASCII characters, and replaces spaces with underscores
func sanitizeFilename(filename string) string {
	// Replace problematic characters and spaces with underscores
	replacer := strings.NewReplacer(
		"/", "",
		"\\", "",
		":", "",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
		"(", "",
		")", "",
		".", "",
	)

	sanitized := replacer.Replace(filename)

	// Convert non-ASCII characters to ASCII equivalents or remove them
	var asciiOnly strings.Builder
	for _, r := range sanitized {
		if r <= 127 { // ASCII range
			asciiOnly.WriteRune(r)
		} else {
			// Simple ASCII conversion for common characters
			switch {
			case r >= 'À' && r <= 'Ÿ': // Latin extended characters
				// Convert accented characters to their base forms
				base := convertAccentedChar(r)
				if base != 0 {
					asciiOnly.WriteRune(base)
				} else {
					asciiOnly.WriteRune(' ')
				}
			default:
				asciiOnly.WriteRune(' ') // Replace other non-ASCII with space
			}
		}
	}
	sanitized = asciiOnly.String()

	// Trim underscores and dots from the ends
	sanitized = strings.Trim(sanitized, "_.")

	// Ensure filename is not empty
	if sanitized == "" {
		sanitized = "unknown"
	}

	sanitized = strings.Join(strings.Fields(sanitized), " ")

	// Limit to 128 characters
	if len(sanitized) > 128 {
		sanitized = sanitized[:128]
		// Ensure we don't end with an underscore after truncation
		sanitized = strings.TrimRight(sanitized, "_")
		if sanitized == "" {
			sanitized = "unknown"
		}
	}

	return sanitized
}

// convertAccentedChar converts common accented characters to their ASCII equivalents
func convertAccentedChar(r rune) rune {
	switch r {
	// A variants
	case 'À', 'Á', 'Â', 'Ã', 'Ä', 'Å':
		return 'A'
	case 'à', 'á', 'â', 'ã', 'ä', 'å':
		return 'a'
	// E variants
	case 'È', 'É', 'Ê', 'Ë':
		return 'E'
	case 'è', 'é', 'ê', 'ë':
		return 'e'
	// I variants
	case 'Ì', 'Í', 'Î', 'Ï':
		return 'I'
	case 'ì', 'í', 'î', 'ï':
		return 'i'
	// O variants
	case 'Ò', 'Ó', 'Ô', 'Õ', 'Ö':
		return 'O'
	case 'ò', 'ó', 'ô', 'õ', 'ö':
		return 'o'
	// U variants
	case 'Ù', 'Ú', 'Û', 'Ü':
		return 'U'
	case 'ù', 'ú', 'û', 'ü':
		return 'u'
	// N variants
	case 'Ñ':
		return 'N'
	case 'ñ':
		return 'n'
	// C variants
	case 'Ç':
		return 'C'
	case 'ç':
		return 'c'
	default:
		return 0 // No ASCII equivalent found
	}
}
