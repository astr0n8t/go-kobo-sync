package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQL query to get highlights after a specific date
const queryStringWithDate = `
SELECT
    b.Text AS Text,
    b.Annotation AS Note,
    c.Title AS Book,
    b.DateCreated AS Timestamp
FROM Bookmark b
JOIN content c ON b.VolumeID = c.ContentID
WHERE b.DateCreated > ?;
`

// SQL query to get all highlights (for initial sync)
const queryStringAll = `
SELECT
    b.Text AS Text,
    b.Annotation AS Note,
    c.Title AS Book,
    b.DateCreated AS Timestamp
FROM Bookmark b
JOIN content c ON b.VolumeID = c.ContentID
`

type Highlight struct {
	Text      *string `json:"text"`
	Note      *string `json:"note,omitempty"`
	Book      *string `json:"book"`
	Timestamp *string `json:"timestamp"`
}

func main() {
	log.Default().Println("Attempting to sync highlights to WebDAV...")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load configuration. Error [%s]\n", err)
	}

	// Create WebDAV client
	webdavClient := NewWebDAVClient(config.WebDAV)

	// Get last sync date
	lastSync, err := webdavClient.GetLastSyncDate()
	if err != nil {
		log.Fatalf("Unable to get last sync date. Error [%s]\n", err)
	}
	log.Printf("Last sync: %s", lastSync)

	// Open Kobo database
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		log.Fatalf("Error opening database. Error [%s]\n", err)
	}
	defer db.Close()

	// Query highlights since last sync
	var rows *sql.Rows
	if lastSync.Unix() == 0 {
		// First sync - get all highlights
		log.Println("First sync - retrieving all highlights")
		rows, err = db.Query(queryStringAll)
	} else {
		// Get highlights since last sync
		lastSyncSplit := strings.Split(lastSync.Format(time.RFC3339), "-")
		lastSyncString := ""
		// Strip timezone information for sqlite query
		if len(lastSyncSplit) > 2 {
			lastSyncString += lastSyncSplit[0] + "-" + lastSyncSplit[1] + "-" + lastSyncSplit[2]
		}
		log.Printf("Retrieving highlights since %s\n", lastSyncString)
		rows, err = db.Query(queryStringWithDate, lastSyncString)
	}

	if err != nil {
		log.Fatalf("Unable to query database. Error [%s]\n", err)
	}
	defer rows.Close()

	// Group highlights by book
	bookHighlights := make(map[string][]Highlight)
	highlightCount := 0

	for rows.Next() {
		var highlight Highlight
		err := rows.Scan(
			&highlight.Text,
			&highlight.Note,
			&highlight.Book,
			&highlight.Timestamp,
		)
		if err != nil {
			log.Fatalf("Unable to scan rows. Error [%s]\n", err)
		}

		// Clean up empty notes
		if highlight.Note != nil && *highlight.Note == "" {
			highlight.Note = nil
		} else if highlight.Note != nil {
			strippedNote := strings.ReplaceAll(*highlight.Note, "\n", " ")
			highlight.Note = &strippedNote
		}

		if highlight.Text != nil {
			strippedHighlight := strings.ReplaceAll(*highlight.Text, "\n", " ")
			highlight.Text = &strippedHighlight
		}

		// Group by book
		if highlight.Book != nil {
			bookTitle := *highlight.Book
			bookHighlights[bookTitle] = append(bookHighlights[bookTitle], highlight)
			highlightCount++
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatalf("Unable to scan rows. Error [%s]\n", err)
	}

	log.Printf("Found %d new highlights across %d books\n", highlightCount, len(bookHighlights))

	numSuccess := 0
	// Process each book
	for bookTitle, highlights := range bookHighlights {
		log.Printf("Processing book: %s (%d highlights)\n", bookTitle, len(highlights))

		// Generate markdown content
		markdownContent, err := GenerateMarkdown(bookTitle, highlights, config.Template)
		if err != nil {
			log.Printf("Error generating markdown for %s: %v\n", bookTitle, err)
			continue
		}

		markdownHeader, err := GenerateMarkdownHeader(bookTitle, config.HeaderTemplate)
		if err != nil {
			log.Printf("Error generating markdown header for %s: %v\n", bookTitle, err)
			continue
		}

		// Save to WebDAV
		err = webdavClient.SaveBookFile(bookTitle, markdownHeader, markdownContent)
		if err != nil {
			log.Printf("Error saving %s to WebDAV: %v\n", bookTitle, err)
			continue
		}

		numSuccess++
		log.Printf("Successfully synced %s\n", bookTitle)
	}

	// Update last sync date
	syncTime := time.Now()
	err = webdavClient.UpdateLastSyncDate(syncTime)
	if err != nil {
		log.Printf("Warning: Could not update last sync date: %v\n", err)
	}

	log.Printf("Sync completed successfully. Processed %d books with %d of %d highlights.\n",
		len(bookHighlights), numSuccess, highlightCount)
}
