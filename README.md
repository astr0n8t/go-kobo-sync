# go-readwise-kobo-sync

Sync [readwise.io](readwise.io) highlights from all books (including sideloaded)
on your Kobo.

## Build

```
git clone git@github.com:isaacgr/go-readwise-kobo-sync.git
cd go-readwise-kobo-sync
make build
```

## Install

Copy the `go-readwise-kobo-sync` directory to the Kobo drive. Make sure to
place it in the `.adds` folder.

## Token

Once installed, update the token file `.adds/go-kobo-readwise-sync/token.txt`
with your readwise.io API token.

If you dont have one, generate it here and then copy it in.

https://readwise.io/access_token

## Nickle Menu

Update your Nicklemenu config with the below:

```
menu_item:main:Sync Highlights to Readwise:cmd_output:5000:/mnt/onboard/.adds/go-readwise-kobo-sync/sync_highlights.sh
```
