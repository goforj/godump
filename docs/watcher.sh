#!/usr/bin/env bash

go install github.com/bokwoon95/wgo@latest

echo "Watching for .go file changes to regenerate documentation..."

wgo -verbose -file=.go -xdir examples \
  go run ./docs/examplegen/main.go :: \
  go run ./docs/readme/main.go
