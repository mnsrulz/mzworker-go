#!/bin/bash

mkdir -p /data
mkdir -p /cache
mkdir -p /rcloneconfig

RCLONE_REMOTE="${RCLONE_REMOTE:?RCLONE_REMOTE env var is required}"
RCLONE_CONFIG="${RCLONE_CONFIG:-/rcloneconfig/rclone.conf}"

rclone mount --allow-non-empty --allow-other --read-only --vfs-read-chunk-size=4M --vfs-cache-max-age=168h \
    --vfs-read-chunk-size-limit=16M --vfs-cache-mode=full --buffer-size=256K --no-checksum \
    --cache-dir=/cache --vfs-cache-max-size=512M "${RCLONE_REMOTE}": /data --config "${RCLONE_CONFIG}" &

while ! mountpoint -q /data; do
    echo "Waiting for Rclone mount..."
    sleep 1
done

echo "Rclone mount is ready."

exec mzworker serve "$@"
