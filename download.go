package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type invoiceLink struct {
	URL  string
	Text string
}

var isoDate = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func buildFileName(firm, rowText, originalName string, now time.Time) string {
	date := isoDate.FindString(rowText)
	if date == "" {
		date = now.Format("2006-01-02")
	}
	cleanFirm := unsafeChars.ReplaceAllString(firm, "_")
	cleanOriginal := unsafeChars.ReplaceAllString(filepath.Base(originalName), "_")
	return fmt.Sprintf("%s_%s_%s", date, cleanFirm, cleanOriginal)
}

func originalNameFrom(resp *http.Response, link invoiceLink) string {
	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err == nil && params["filename"] != "" {
		return params["filename"]
	}
	parsed, err := url.Parse(link.URL)
	if err == nil && parsed.Query().Get("sys_id") != "" {
		return parsed.Query().Get("sys_id") + ".pdf"
	}
	return "invoice.pdf"
}

func downloadInvoice(client *http.Client, link invoiceLink, outDir, firm string, now time.Time) (string, bool, error) {
	resp, err := client.Get(link.URL)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("download %s: status %d", link.URL, resp.StatusCode)
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return "", false, fmt.Errorf("download %s: got an HTML page, the session may have expired", link.URL)
	}
	path := filepath.Join(outDir, buildFileName(firm, link.Text, originalNameFrom(resp, link), now))
	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(path)
		return "", false, err
	}
	return path, true, nil
}

func defaultFirm(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return "invoices"
	}
	return strings.Split(parsed.Hostname(), ".")[0]
}
