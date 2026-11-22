package main

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)

func TestBuildFileName(t *testing.T) {
	cases := []struct {
		name, rowText, original, want string
	}{
		{"date in row", "INV-1 2026-03-15 $100", "invoice 1.pdf", "2026-03-15_acme_invoice_1.pdf"},
		{"fallback to now", "INV-2", "a.pdf", "2026-11-22_acme_a.pdf"},
		{"path traversal", "", "../../etc/passwd", "2026-11-22_acme_passwd"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildFileName("acme", c.rowText, c.original, fixedNow); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestDefaultFirm(t *testing.T) {
	if got := defaultFirm("https://acme.service-now.com/cts?id=invoices"); got != "acme" {
		t.Fatalf("got %q", got)
	}
}
