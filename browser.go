package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

const (
	emailSelector    = chromedp.CSS(`input[type="email"], input[name="user_name"], input#user_name, input[type="text"]`)
	passwordSelector = chromedp.CSS(`input[type="password"]`)
	linkSelector     = chromedp.CSS(`a[href*="sys_attachment.do"]`)
	collectLinksJS   = `Array.from(document.querySelectorAll('a[href*="sys_attachment.do"]')).map(a => ({
		URL: a.href,
		Text: (a.closest('tr, li, .panel') || a.parentElement).innerText
	}))`
)

func fetchInvoiceLinks(parent context.Context, pageURL, email, password string, headed bool) ([]invoiceLink, []*http.Cookie, error) {
	options := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", !headed))
	if os.Getenv("CHROME_NO_SANDBOX") != "" {
		options = append(options, chromedp.NoSandbox)
	}
	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(parent, options...)
	defer cancelAllocator()
	browserCtx, cancelBrowser := chromedp.NewContext(allocatorCtx)
	defer cancelBrowser()

	if err := chromedp.Do(browserCtx, chromedp.Navigate(pageURL)); err != nil {
		return nil, nil, err
	}
	linksWait := 30 * time.Second
	if headed {
		linksWait = 5 * time.Minute
	}
	timeoutCtx, cancelTimeout := context.WithTimeout(browserCtx, linksWait+2*time.Minute)
	defer cancelTimeout()
	if !headed {
		if err := login(timeoutCtx, email, password); err != nil {
			return nil, nil, err
		}
		if err := chromedp.Do(timeoutCtx, chromedp.Navigate(pageURL)); err != nil {
			return nil, nil, err
		}
	}
	linksCtx, cancelLinks := context.WithTimeout(timeoutCtx, linksWait)
	defer cancelLinks()
	if err := chromedp.Do(linksCtx, chromedp.WaitVisible(linkSelector)); err != nil {
		saveDebugPage(browserCtx)
		return nil, nil, errors.New("no invoice links found, page saved to debug-page.html")
	}
	links, err := chromedp.Run(timeoutCtx, chromedp.Evaluate[[]invoiceLink](collectLinksJS))
	if err != nil {
		return nil, nil, err
	}
	cookies, err := browserCookies(timeoutCtx, pageURL)
	return links, cookies, err
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

func browserCookies(ctx context.Context, pageURL string) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	err := chromedp.Do(ctx, chromedp.Func(func(ctx context.Context, target *chromedp.Target) error {
		result, err := cdp.Call(ctx, target, network.GetCookies, network.GetCookiesParams{URLs: []string{pageURL}})
		for _, cookie := range result.Cookies {
			cookies = append(cookies, &http.Cookie{Name: cookie.Name, Value: cookie.Value, Path: cookie.Path})
		}
		return err
	}))
	return cookies, err
}

func saveDebugPage(ctx context.Context) {
	html, err := chromedp.Run(ctx, chromedp.Evaluate[string](`location.href + "\n" + document.documentElement.outerHTML`))
	if err == nil {
		os.WriteFile("debug-page.html", []byte(html), 0o644)
	}
}
