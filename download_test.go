package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestDownloadInvoiceSendsCookieAndSkipsExisting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("session"); err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Disposition", `attachment; filename="bill.pdf"`)
		w.Write([]byte("pdf-bytes"))
	}))
	defer server.Close()

	jar := newTestJar(t, server.URL)
	client := &http.Client{Jar: jar}
	outDir := t.TempDir()
	link := invoiceLink{URL: server.URL + "/sys_attachment.do?sys_id=1", Text: "2026-01-02"}

	path, written, err := downloadInvoice(client, link, outDir, "fx", fixedNow)
	if err != nil || !written {
		t.Fatalf("first download: written=%v err=%v", written, err)
	}
	if filepath.Base(path) != "2026-01-02_fx_bill.pdf" {
		t.Fatalf("unexpected name %q", path)
	}
	content, _ := os.ReadFile(path)
	if string(content) != "pdf-bytes" {
		t.Fatalf("unexpected content %q", content)
	}
	if _, written, err = downloadInvoice(client, link, outDir, "fx", fixedNow); err != nil || written {
		t.Fatalf("second download: written=%v err=%v", written, err)
	}
}

func TestDownloadInvoiceRejectsForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()
	if _, _, err := downloadInvoice(http.DefaultClient, invoiceLink{URL: server.URL}, t.TempDir(), "fx", fixedNow); err == nil {
		t.Fatal("expected an error")
	}
}
