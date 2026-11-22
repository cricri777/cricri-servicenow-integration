package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

const (
	emailSelector    = chromedp.CSS(`input[type="email"], input[name="user_name"], input#user_name, input[type="text"]`)
	passwordSelector = chromedp.CSS(`input[type="password"]`)
	rowCSS           = `tr[ng-click^="onInvoiceClick"]`
	rowSelector      = chromedp.CSS(rowCSS)
	downloadTimeout  = time.Minute
)

func downloadInvoices(parent context.Context, cfg config) error {
	options := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", !cfg.Headed))
	if os.Getenv("CHROME_NO_SANDBOX") != "" {
		options = append(options, chromedp.NoSandbox)
	}
	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(parent, options...)
	defer cancelAllocator()
	browserCtx, cancelBrowser := chromedp.NewContext(allocatorCtx)
	defer cancelBrowser()

	if err := chromedp.Do(browserCtx, chromedp.Navigate(cfg.URL)); err != nil {
		return err
	}
	rowsWait := 30 * time.Second
	if cfg.Headed {
		rowsWait = 5 * time.Minute
	}
	waitCtx, cancelWait := context.WithTimeout(browserCtx, rowsWait+2*time.Minute)
	defer cancelWait()
	if !cfg.Headed {
		if err := login(waitCtx, cfg.Email, cfg.Password); err != nil {
			return err
		}
		if err := chromedp.Do(waitCtx, chromedp.Navigate(cfg.URL)); err != nil {
			return err
		}
	}
	rowsCtx, cancelRows := context.WithTimeout(waitCtx, rowsWait)
	defer cancelRows()
	if err := chromedp.Do(rowsCtx, chromedp.WaitVisible(rowSelector)); err != nil {
		saveDebugPage(browserCtx)
		return errors.New("no invoice rows found, page saved to debug-page.html")
	}

	tempDir, err := os.MkdirTemp(cfg.OutDir, ".download-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := allowDownloads(browserCtx, tempDir); err != nil {
		return err
	}
	rowCount, err := chromedp.Run(browserCtx, chromedp.Evaluate[int](fmt.Sprintf(`document.querySelectorAll(%q).length`, rowCSS)))
	if err != nil {
		return err
	}
	var failures []error
	for index := 0; index < rowCount; index++ {
		path, written, err := downloadRow(browserCtx, index, tempDir, cfg)
		if err != nil {
			failures = append(failures, fmt.Errorf("row %d: %w", index+1, err))
			continue
		}
		if written {
			fmt.Println("downloaded", path)
		} else {
			fmt.Println("skipped", path)
		}
	}
	return errors.Join(failures...)
}

func allowDownloads(ctx context.Context, directory string) error {
	return chromedp.Do(ctx, chromedp.Func(func(ctx context.Context, target *chromedp.Target) error {
		_, err := cdp.Call(ctx, target, browser.SetDownloadBehavior, browser.SetDownloadBehaviorParams{
			Behavior:     browser.SetDownloadBehaviorBehaviorAllow,
			DownloadPath: directory,
		})
		return err
	}))
}

func downloadRow(ctx context.Context, index int, tempDir string, cfg config) (string, bool, error) {
	rowCtx, cancelRow := context.WithTimeout(ctx, downloadTimeout)
	defer cancelRow()
	rowDir := filepath.Join(tempDir, fmt.Sprint(index))
	if err := os.Mkdir(rowDir, 0o755); err != nil {
		return "", false, err
	}
	if err := allowDownloads(rowCtx, rowDir); err != nil {
		return "", false, err
	}
	row := fmt.Sprintf(`document.querySelectorAll(%q)[%d]`, rowCSS, index)
	rowText, err := chromedp.Run(rowCtx, chromedp.Evaluate[string](row+`.innerText`))
	if err != nil {
		return "", false, err
	}
	if err := chromedp.Do(rowCtx, chromedp.Evaluate[chromedp.Void](row+`.click()`)); err != nil {
		return "", false, err
	}
	downloaded, err := waitForDownload(rowCtx, rowDir)
	if err != nil {
		return "", false, err
	}
	return saveDownload(downloaded, rowText, cfg)
}

func waitForDownload(ctx context.Context, directory string) (string, error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".crdownload") {
				return filepath.Join(directory, entry.Name()), nil
			}
		}
		select {
		case <-ctx.Done():
			return "", errors.New("no download finished after the click")
		case <-ticker.C:
		}
	}
}

func login(ctx context.Context, email, password string) error {
	probeCtx, cancelProbe := context.WithTimeout(ctx, 15*time.Second)
	defer cancelProbe()
	if err := chromedp.Do(probeCtx, chromedp.WaitVisible(passwordSelector)); err != nil {
		return nil
	}
	return chromedp.Do(ctx,
		chromedp.SendKeys(emailSelector, email),
		chromedp.SendKeys(passwordSelector, password+kb.Enter),
		chromedp.Sleep(5*time.Second),
	)
}

func saveDebugPage(ctx context.Context) {
	html, err := chromedp.Run(ctx, chromedp.Evaluate[string](`location.href + "\n" + document.documentElement.outerHTML`))
	if err == nil {
		os.WriteFile("debug-page.html", []byte(html), 0o644)
	}
}
