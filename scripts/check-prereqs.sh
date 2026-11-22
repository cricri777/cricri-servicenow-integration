#!/usr/bin/env bash
set -euo pipefail

missing=0

if ! command -v go >/dev/null 2>&1; then
  echo "missing prerequisite: go (https://go.dev/dl/)" >&2
  missing=1
fi

chrome_found=0
for binary in google-chrome google-chrome-stable chromium chromium-browser; do
  if command -v "$binary" >/dev/null 2>&1; then
    chrome_found=1
    break
  fi
done
if [ "$chrome_found" -eq 0 ]; then
  echo "missing prerequisite: Chrome or Chromium (google-chrome, google-chrome-stable, chromium or chromium-browser)" >&2
  missing=1
fi

if [ "${1:-}" = "--env" ]; then
  for variable in SN_EMAIL SN_PASS; do
    if [ -z "${!variable:-}" ]; then
      echo "missing environment variable: $variable" >&2
      missing=1
    fi
  done
fi

exit "$missing"
