#!/usr/bin/env bash
set -o errexit

mkdir -p bin

if [ ! -f bin/yt-dlp ]; then
  echo "Downloading yt-dlp..."
  curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o bin/yt-dlp
  chmod a+rx bin/yt-dlp
fi

go build -tags netgo -ldflags '-s -w' -o app main.go
