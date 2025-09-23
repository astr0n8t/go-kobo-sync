# go-kobo-sync

Sync Kobo highlights to a WebDAV server in markdown format. Minimizes writes to the Kobo flash storage by only syncing highlights created since the last sync.

## Features

- **WebDAV Integration**: Sync highlights to any WebDAV-compatible server (Nextcloud, ownCloud, etc.)
- **Incremental Sync**: Only syncs highlights created since the last sync to minimize Kobo flash writes
- **Markdown Output**: Stores highlights in organized markdown files, one per book
- **Template Support**: Customizable markdown templates for highlight formatting
- **Atomic Operations**: Safe file operations with backup and atomic moves to prevent corruption
- **HTTPS Support**: Built-in CA certificate support for secure connections

## Build

```bash
git clone https://github.com/astr0n8t/go-kobo-sync.git
cd go-kobo-sync
make build
```

## Install

1. Copy the `go-kobo-sync` directory to the Kobo drive in the `.adds` folder
2. Configure your WebDAV settings (see Configuration section)

## Configuration

### WebDAV Settings

Copy `config.example.txt` to `config.txt` and update with your WebDAV server details:

```
# WebDAV server URL (required)
webdav_url=https://your-nextcloud.com/remote.php/dav/files/username

# WebDAV credentials (required)  
webdav_username=your_username
webdav_password=your_password

# Base path for storing highlights (optional, defaults to /kobo-highlights)
webdav_path=/kobo-highlights
```

### Template Customization (Optional)

Copy `template.example.md` to `template.md` to customize the markdown format. The template uses Go's text/template syntax with the following variables:

- `{{.Title}}` - Book title
- `{{.SyncDate}}` - Last sync timestamp  
- `{{.Highlights}}` - Array of highlights with:
  - `{{.Text}}` - Highlight text
  - `{{.Note}}` - Your annotation/note
  - `{{.Timestamp}}` - When the highlight was created

## Nickle Menu Integration

Add this to your Nicklemenu config:

```
menu_item:main:Sync Highlights to WebDAV:cmd_output:5000:/mnt/onboard/.adds/go-kobo-sync/sync_highlights.sh
```

## How It Works

1. **First Run**: Syncs all existing highlights and creates a `last_sync.txt` file on your WebDAV server
2. **Subsequent Runs**: Only syncs highlights created since the last sync date
3. **Per-Book Files**: Creates/updates one markdown file per book on your WebDAV server
4. **Safe Updates**: Downloads existing files, merges new highlights, and atomically updates with backup

## File Structure on WebDAV Server

```
/kobo-highlights/
├── last_sync.txt              # Tracks last sync date
├── Book_Title_1.md            # Highlights for book 1
├── Another_Book_Title.md      # Highlights for book 2
└── ...
```

## HTTPS Certificate Support

Place your CA certificates in the `ca-certs/` directory for HTTPS WebDAV servers. Let's Encrypt certificates will be added in a future update.
