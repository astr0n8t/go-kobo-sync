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
const defaultTemplate = `

{{range .Highlights}}
---
**{{.Timestamp}}**
{{if .Text}}> {{.Text}}{{end}}
{{if .Note}}*{{.Note}}*{{end}}
{{end}}`

const defaultHeaderTemplate = `# {{.Title}}

`

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

	for i := range len(highlights) {
		timestamp, err := time.Parse("2006-01-02T15:04:05.000", *highlights[i].Timestamp)
		if err != nil {
			continue
		}

		formattedTimestamp := timestamp.Format(time.RFC1123)
		highlights[i].Timestamp = &formattedTimestamp
	}

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

func GenerateMarkdownHeader(bookTitle string, templateStr string) ([]byte, error) {
	if templateStr == "" {
		templateStr = defaultHeaderTemplate
	}

	bookHighlights := BookHighlights{
		Title:    bookTitle,
		SyncDate: time.Now(),
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
