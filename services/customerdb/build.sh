#!/bin/sh
sed -i '4,$d' go.mod
docker build -t tft/customerdb:latest .
git restore go.mod