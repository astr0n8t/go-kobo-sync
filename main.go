package main

import (
	"database/sql"
	"log"
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
WHERE b.DateCreated > '?';
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

// TODO: Add more fields, or some ways to set what type of highlight this is
// (i.e. book, quote etc.)
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
		log.Fatalf("Unable to load configuration. Error [%s]", err)
	}

	// Create WebDAV client
	webdavClient := NewWebDAVClient(config.WebDAV)

	// Get last sync date
	lastSync, err := webdavClient.GetLastSyncDate()
	if err != nil {
		log.Fatalf("Unable to get last sync date. Error [%s]", err)
	}
	log.Printf("Last sync: %s", lastSync.Format(time.RFC3339))

	// Open Kobo database
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		log.Fatalf("Error opening database. Error [%s]", err)
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
		log.Printf("Retrieving highlights since %s", lastSync.Format(time.RFC3339))
		rows, err = db.Query(queryStringWithDate, lastSync.Format(time.RFC3339))
	}

	if err != nil {
		log.Fatalf("Unable to query database. Error [%s]", err)
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
			log.Fatalf("Unable to scan rows. Error [%s]", err)
		}

		// Clean up empty notes
		if highlight.Note != nil && *highlight.Note == "" {
			highlight.Note = nil
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
		log.Fatalf("Unable to scan rows. Error [%s]", err)
	}

	log.Printf("Found %d new highlights across %d books", highlightCount, len(bookHighlights))

	// Process each book
	for bookTitle, highlights := range bookHighlights {
		log.Printf("Processing book: %s (%d highlights)", bookTitle, len(highlights))

		// Get existing file content
		existingContent, err := webdavClient.GetBookFile(bookTitle)
		if err != nil {
			log.Printf("Warning: Could not retrieve existing file for %s: %v", bookTitle, err)
		}

		var allHighlights []Highlight
		if len(existingContent) > 0 {
			// Parse existing highlights to avoid duplicates
			existingHighlights := ParseExistingMarkdown(existingContent)
			allHighlights = MergeHighlights(existingHighlights, highlights)
			log.Printf("Merged %d existing highlights with %d new highlights for %s",
				len(existingHighlights), len(highlights), bookTitle)
		} else {
			allHighlights = highlights
		}

		// Generate markdown content
		markdownContent, err := GenerateMarkdown(bookTitle, allHighlights, config.Template)
		if err != nil {
			log.Printf("Error generating markdown for %s: %v", bookTitle, err)
			continue
		}

		// Save to WebDAV
		err = webdavClient.SaveBookFile(bookTitle, markdownContent)
		if err != nil {
			log.Printf("Error saving %s to WebDAV: %v", bookTitle, err)
			continue
		}

		log.Printf("Successfully synced %s", bookTitle)
	}

	// Update last sync date
	syncTime := time.Now()
	err = webdavClient.UpdateLastSyncDate(syncTime)
	if err != nil {
		log.Printf("Warning: Could not update last sync date: %v", err)
	}

	log.Printf("Sync completed successfully. Processed %d books with %d highlights.",
		len(bookHighlights), highlightCount)
}
