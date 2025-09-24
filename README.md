# go-kobo-sync

Sync Kobo highlights to a WebDAV server in markdown format. Only syncs highlights created since the last sync.

## Features

- **WebDAV Integration**: Sync highlights to any WebDAV-compatible server (Nextcloud, ownCloud, etc.)
- **Incremental Sync**: Only syncs highlights created since the last sync
- **Markdown Output**: Stores highlights in organized markdown files, one per book
- **Template Support**: Customizable markdown templates for highlight formatting
- **HTTPS Support**: Built-in CA certificate support for secure connections

## Build

```bash
git clone https://github.com/astr0n8t/go-kobo-sync.git
cd go-kobo-sync
make docker-build
```

## Install

1. Configure your WebDAV settings (see Configuration section)

2. Push to your device:
```
# you may need to edit the Makefile with the path to your Kobo
make install

```


## Configuration

### WebDAV Settings

Defaults:
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

`header_template.md` is used to create the file initially and `template.md` is used to add highlights to the file. The template uses Go's text/template syntax with the following variables:

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

1. **First Run**: Syncs all existing highlights and creates a `.last_sync` file on your WebDAV server
2. **Subsequent Runs**: Only syncs highlights created since the last sync date
3. **Per-Book Files**: Creates/updates one markdown file per book on your WebDAV server
4. **Safe Updates**: Downloads existing files, merges new highlights, and atomically updates with backup

## File Structure on WebDAV Server

```
/kobo-highlights/
├── .last_sync                 # Tracks last sync date
├── Book_Title_1.md            # Highlights for book 1
├── Another_Book_Title.md      # Highlights for book 2
└── ...
```

## HTTPS Certificate Support

Place your CA certificates in the `ca-certs/` directory for HTTPS WebDAV servers. By default, the Mozilla CA bundle is grabbed with `make certs` in order to support Let's Encrypt.

