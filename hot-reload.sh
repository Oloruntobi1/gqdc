#!/bin/sh
set -e

CompileDaemon \
-log-prefix=false \
-exclude-dir=".git" \
-exclude-dir=".github" \
-build="go build -a -installsuffix cgo -o ./build/web cmd/web/main.go" \
-graceful-kill=true \
-graceful-timeout=15 \
-command="./build/web" -verbose
