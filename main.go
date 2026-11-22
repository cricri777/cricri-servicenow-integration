package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

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
	flag.BoolVar(&cfg.Headed, "headed", false, "log in by hand in a visible browser window")
	flag.Parse()
	if !cfg.Headed {
		if err := promptMissingCredentials(&cfg); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
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
	if !cfg.Headed && (cfg.Email == "" || cfg.Password == "") {
		return errors.New("SN_EMAIL and SN_PASS must be set")
	}
	if cfg.Firm == "" {
		cfg.Firm = defaultFirm(cfg.URL)
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return err
	}
	return downloadInvoices(ctx, cfg)
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
