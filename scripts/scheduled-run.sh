#!/usr/bin/env bash
set -uo pipefail

notify() {
  curl -fsS -H "Title: sn-invoices" -d "$1" "https://ntfy.sh/$NTFY_TOPIC" >/dev/null
}

cd "$(dirname "$0")/.."

if ! make run; then
  notify "make run failed on $(date +%F)"
  exit 1
fi

if [ "$(date +%d)" -ge 21 ] && [ -z "$(find "$SN_OUT" -maxdepth 1 -type f -mtime -31 -print -quit)" ]; then
  notify "No invoice was downloaded in the last 31 days"
  exit 1
fi
