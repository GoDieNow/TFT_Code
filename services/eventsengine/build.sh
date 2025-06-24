#!/bin/sh
sed -i '4,$d' go.mod
docker build -t tft/eventsengine:latest .
git restore go.mod