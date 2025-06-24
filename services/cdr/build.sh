#!/bin/sh
sed -i '4,$d' go.mod
docker build -t tft/cdr:latest .
git restore go.mod