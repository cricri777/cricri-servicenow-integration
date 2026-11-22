package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func newTestJar(t *testing.T, rawURL string) *cookiejar.Jar {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(rawURL)
	jar.SetCookies(parsed, []*http.Cookie{{Name: "session", Value: "ok"}})
	return jar
}

func fakePortal() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/portal", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := r.Cookie("session"); err != nil {
			fmt.Fprint(w, `<form method="post" action="/login">
				<input type="text" name="user_name"><input type="password" name="pass"><input type="submit"></form>`)
			return
		}
		fmt.Fprint(w, `<table>
			<tr><td>INV-1 2026-01-10</td><td><a href="/sys_attachment.do?sys_id=a">a</a></td></tr>
			<tr><td>INV-2 2026-02-10</td><td><a href="/sys_attachment.do?sys_id=b">b</a></td></tr></table>`)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("user_name") != "me@example.com" || r.FormValue("pass") != "secret" {
			http.Error(w, "bad credentials", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "ok", Path: "/"})
		http.Redirect(w, r, "/portal", http.StatusSeeOther)
	})
	mux.HandleFunc("/sys_attachment.do", func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("session"); err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, r.URL.Query().Get("sys_id")))
		fmt.Fprint(w, "pdf")
	})
	return mux
}

func TestEndToEndAgainstFakePortal(t *testing.T) {
	if _, err := exec.LookPath("chromium"); err != nil {
		if _, err := exec.LookPath("google-chrome"); err != nil {
			if _, err := exec.LookPath("google-chrome-stable"); err != nil {
				t.Skip("no Chrome or Chromium installed")
			}
		}
	}
	server := httptest.NewServer(fakePortal())
	defer server.Close()
	outDir := t.TempDir()
	cfg := config{URL: server.URL + "/portal", OutDir: outDir, Firm: "fx", Email: "me@example.com", Password: "secret"}

	if err := run(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"2026-01-10_fx_a.pdf", "2026-02-10_fx_b.pdf"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	if err := run(context.Background(), cfg); err != nil {
		t.Fatalf("second run: %v", err)
	}
	entries, _ := os.ReadDir(outDir)
	if len(entries) != 2 {
		t.Fatalf("expected 2 files after rerun, got %d", len(entries))
	}
}
