#!/bin/sh
sed -i '4,$d' go.mod
docker build -t tft/planmanager:latest .
git restore go.mod