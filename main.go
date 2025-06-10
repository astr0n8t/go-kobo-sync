package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: Pagination
const queryString = `
SELECT
    Text as Text,
    Annotation as Note,
    VolumeID as Book,
    DateCreated as Timestamp
FROM Bookmark
WHERE Type = 'highlight';
`
const tokenPath = "/mnt/onboard/.adds/go-readwise-kobo-sync/token.txt"
const dbLocation = "/mnt/onboard/.kobo/KoboReader.sqlite"

// TODO: Add more fields, or some ways to set what type of highlight this is
// (i.e. book, quote etc.)
type Highlight struct {
	Text      *string `json:"text"`
	Note      *string `json:"note,omitempty"`
	Book      *string `json:"book"`
	Timestamp *string `json:"timestamp"`
}

type HighlightPost struct {
	Highlights []Highlight `json:"highlights"`
}

func main() {
	log.Default().Println("Attempting to sync...")

	// readwise.io/access_token
	tokenData, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalf("Unable to read API token. Error [%s]", err)
	}

	// Sanitize the token after reading it in.
	// I had this problem when trying to read from file from the kobo but
	// not when reading from my PC.
	// Adding this fixed it, maybe user error but this cant hurt really.
	token := strings.TrimSpace(string(tokenData))

	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		log.Fatalf("Error opening database. Error [%s]", err)
	}

	defer db.Close()

	rows, err := db.Query(queryString)

	if err != nil {
		log.Fatalf("Unable to query database. Error [%s]", err)
	}
	defer rows.Close()

	var higlights []Highlight

	for rows.Next() {
		var higlight Highlight
		err := rows.Scan(
			&higlight.Text,
			&higlight.Note,
			&higlight.Book,
			&higlight.Timestamp,
		)
		if err != nil {
			log.Fatalf("Unable to scan rows. Error [%s]", err)
		}
		if higlight.Note != nil && *higlight.Note == "" {
			higlight.Note = nil
		}
		higlights = append(higlights, higlight)
	}

	err = rows.Err()
	if err != nil {
		log.Fatalf("Unable to scan rows. Error [%s]", err)
	}

	higlightPost := &HighlightPost{Highlights: higlights}

	body, err := json.Marshal(higlightPost)

	if err != nil {
		log.Fatalf("Unable to convert higlights to json. Error [%s]", err)
	}

	// https://readwise.io/api_deets
	req, err := http.NewRequest(
		"POST",
		"https://readwise.io/api/v2/highlights/",
		bytes.NewBuffer(body),
	)

	var tokenHeader = "Token " + token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", tokenHeader)

	if err != nil {
		log.Fatalf("Unable to build request. Error [%s]", err)
	}

	client := GetClient()
	res, err := client.Do(req)

	if err != nil {
		log.Fatalf("Unable to post higlights. Error [%s]", err)
	}

	if res.StatusCode == http.StatusOK {
		log.Default().Println("Readwise highlights synced.")
	} else {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf(
				"Failure syncing readwise highlights, "+
					"however I am unable to parse response body. Error [%s]",
				err,
			)
		}
		log.Fatalf(
			"Failure syncing readwise higlights. Error [%s]",
			string(bodyBytes),
		)
	}
}
