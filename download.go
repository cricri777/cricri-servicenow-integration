package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

func saveDownload(downloadedPath, rowText string, cfg config) (string, bool, error) {
	path := filepath.Join(cfg.OutDir, buildFileName(cfg.Firm, rowText, filepath.Base(downloadedPath), time.Now()))
	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	}
	if err := os.Rename(downloadedPath, path); err != nil {
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
