#!/bin/sh

# Based on trial and error, nicklemenu requires that the output of the script
# is captured through an echo, or else it just fails with various bash
# codes (i.e. exit 1, exit 127) without explaining why
echo $(./mnt/onboard/.adds/go-readwise-kobo-sync/sync_highlights)