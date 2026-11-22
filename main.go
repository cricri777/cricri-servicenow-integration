package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type config struct {
	URL      string
	OutDir   string
	Firm     string
	Headed   bool
	Email    string
	Password string
}

func main() {
	cfg := config{Email: os.Getenv("SN_EMAIL"), Password: os.Getenv("SN_PASS")}
	flag.StringVar(&cfg.URL, "url", "", "invoices page URL")
	flag.StringVar(&cfg.OutDir, "out", "", "output directory")
	flag.StringVar(&cfg.Firm, "name", "", "firm name used in file names")
	flag.BoolVar(&cfg.Headed, "headed", false, "show the browser window")
	flag.Parse()
	if err := promptMissingCredentials(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := run(context.Background(), cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config) error {
	if cfg.URL == "" || cfg.OutDir == "" {
		return errors.New("--url and --out are required")
	}
	if cfg.Email == "" || cfg.Password == "" {
		return errors.New("SN_EMAIL and SN_PASS must be set")
	}
	if cfg.Firm == "" {
		cfg.Firm = defaultFirm(cfg.URL)
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return err
	}
	links, cookies, err := fetchInvoiceLinks(ctx, cfg.URL, cfg.Email, cfg.Password, cfg.Headed)
	if err != nil {
		return err
	}
	pageURL, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	jar.SetCookies(pageURL, cookies)
	client := &http.Client{Jar: jar, Timeout: 2 * time.Minute}
	var failures []error
	for _, link := range links {
		path, written, err := downloadInvoice(client, link, cfg.OutDir, cfg.Firm, time.Now())
		if err != nil {
			failures = append(failures, err)
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

func promptMissingCredentials(cfg *config) error {
	if cfg.Email == "" {
		fmt.Fprint(os.Stderr, "Email: ")
		email, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		cfg.Email = strings.TrimSpace(email)
	}
	if cfg.Password == "" {
		fmt.Fprint(os.Stderr, "Password: ")
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		cfg.Password = string(password)
	}
	return nil
}
