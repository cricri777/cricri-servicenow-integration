#!/usr/bin/env bash
set -euo pipefail

unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
  echo "gofmt needed:" >&2
  echo "$unformatted" >&2
  exit 1
fi

go vet ./...
go test -v ./...
