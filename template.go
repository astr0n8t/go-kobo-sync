package main

import (
	"bytes"
	"sort"
	"text/template"
	"time"
)

// BookHighlights represents all highlights for a specific book
type BookHighlights struct {
	Title      string
	Highlights []Highlight
	SyncDate   time.Time
}

// Default markdown template for book highlights
const defaultTemplate = `# {{.Title}}

*Synced on {{.SyncDate.Format "2006-01-02 15:04:05"}}*

{{range .Highlights}}
---

**{{.Timestamp}}**

{{if .Text}}> {{.Text}}{{end}}

{{if .Note}}*{{.Note}}*{{end}}

{{end}}`

// GenerateMarkdown creates markdown content for a book's highlights using a template
func GenerateMarkdown(bookTitle string, highlights []Highlight, templateStr string) ([]byte, error) {
	if templateStr == "" {
		templateStr = defaultTemplate
	}

	// Sort highlights by timestamp
	sort.Slice(highlights, func(i, j int) bool {
		if highlights[i].Timestamp == nil || highlights[j].Timestamp == nil {
			return false
		}
		// Parse timestamps and compare
		t1, err1 := time.Parse("2006-01-02T15:04:05.000", *highlights[i].Timestamp)
		t2, err2 := time.Parse("2006-01-02T15:04:05.000", *highlights[j].Timestamp)
		if err1 != nil || err2 != nil {
			return false
		}
		return t1.Before(t2)
	})

	bookHighlights := BookHighlights{
		Title:      bookTitle,
		Highlights: highlights,
		SyncDate:   time.Now(),
	}

	tmpl, err := template.New("book").Parse(templateStr)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, bookHighlights)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MergeHighlights merges existing highlights with new ones, avoiding duplicates
func MergeHighlights(existing []Highlight, new []Highlight) []Highlight {
	// Create a map of existing highlights for deduplication
	existingMap := make(map[string]bool)
	for _, h := range existing {
		key := generateHighlightKey(h)
		existingMap[key] = true
	}

	// Add new highlights that don't already exist
	var merged []Highlight
	merged = append(merged, existing...)

	for _, h := range new {
		key := generateHighlightKey(h)
		if !existingMap[key] {
			merged = append(merged, h)
		}
	}

	return merged
}

// generateHighlightKey creates a unique key for a highlight to detect duplicates
func generateHighlightKey(h Highlight) string {
	var key string
	if h.Text != nil {
		key += *h.Text
	}
	key += "|"
	if h.Note != nil {
		key += *h.Note
	}
	key += "|"
	if h.Timestamp != nil {
		key += *h.Timestamp
	}
	return key
}

// ParseExistingMarkdown extracts highlights from existing markdown content
// This is a simple parser - could be enhanced for more complex markdown
func ParseExistingMarkdown(content []byte) []Highlight {
	var highlights []Highlight

	// Simple parsing - this would need to be more sophisticated
	// for complex markdown structures. For now, we'll assume the
	// format matches our template output.

	// TODO: Implement proper markdown parsing if needed
	// For now, return empty slice to avoid duplicate highlights
	// The WebDAV merge strategy will handle this by appending new highlights

	return highlights
}
