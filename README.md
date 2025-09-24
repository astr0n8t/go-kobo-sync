# go-kobo-sync

Sync Kobo highlights to a WebDAV server in markdown format. Only syncs highlights created since the last sync.

## Motivation

I use this in conjunction with the [Obsidian Remotely Save](https://github.com/remotely-save/remotely-save) plugin to sync my Kobo annotations into my Obsidian vault using WebDAV.  Since it generates Markdown files, you could use it with whatever you wanted.  If you don't want to use WebDAV, you should be able to use this repo as a stepping stone to implementing whatever sync technology you wanted like I did originally.

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

### NickelMenu Install (Optional)

You can add options to [NickelMenu](https://pgaskin.net/NickelMenu/) that allow you to manually trigger syncing and to manage the auto-sync hook.

### AutoSyncing (Optional)

If you want to sync every time the device connects to WiFi, simply SSH on and run the following command on your Kobo or select install on your NickelMenu:
```
cp /mnt/onboard/.adds/go-kobo-sync/97-synchighlights.rules /etc/udev/rules.d/
```

Similarly to uninstall the auto sync hook just remove the file from udev.

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


## Debugging

If for some reason you run into any issues, you can enable logging in the `exec_on_wifi_add.sh` script and then logs will be written to `/mnt/onboard/.adds/go-kobo-sync/log` on every attempted run.

If you need to reset and re-sync everything, just remove the `.last_sync` file from your WebDAV server.  The sync is append only so after a sync has been made and you modify the respective file generated the annotation will not come back but new annotations for that book will be appended to that file.

If you run into any issues please open an issue but know that I consider this extremely beta software and don't have a lot of free time to support it.
