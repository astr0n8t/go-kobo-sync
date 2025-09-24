#!/bin/sh


# UNCOMMENT TO ENABLE LOGFILE
#exec 1>> /mnt/onboard/.adds/go-kobo-sync/log 2>&1

echo $(date) wlan0 added attempting to sync highlights
# Add to command to enable logging
# >> /mnt/onboard/.adds/go-kobo-sync/log 2>&1 \
{ sh -c 'sleep 5 && /mnt/onboard/.adds/go-kobo-sync/sync_highlights' & } &
