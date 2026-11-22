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
		{"date in row", "INV-1 2026-03-15 $100", "invoice 1.pdf", "2026-03-15_fx_invoice_1.pdf"},
		{"fallback to now", "INV-2", "a.pdf", "2026-11-22_fx_a.pdf"},
		{"path traversal", "", "../../etc/passwd", "2026-11-22_fx_passwd"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildFileName("fx", c.rowText, c.original, fixedNow); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestDefaultFirm(t *testing.T) {
	if got := defaultFirm("https://fx.service-now.com/cts?id=invoices"); got != "fx" {
		t.Fatalf("got %q", got)
	}
}
