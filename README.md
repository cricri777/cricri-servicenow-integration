# cricri-servicenow-integration

Downloads the invoices from a ServiceNow portal into a local directory.

This tool is custom for one client portal and one firm. It reads the invoice rows of that portal page and clicks each row, so it will not work as is on other ServiceNow portals.

## Set up a new laptop

Linux only.

1. Install Go (https://go.dev/dl/), Chrome or Chromium, `git` and `make`.
2. Clone the repo and check the prerequisites.

```
git clone https://github.com/cricri777/cricri-servicenow-integration.git
cd cricri-servicenow-integration
make check
```

3. Store the credentials in a file outside the repo that only you can read.

```
mkdir -p ~/.config/sn-invoices
touch ~/.config/sn-invoices/env && chmod 600 ~/.config/sn-invoices/env
```

Then put these lines in `~/.config/sn-invoices/env`. The `SN_URL`, `SN_OUT` and `SN_NAME` lines are optional defaults for `make run`, which sees them only after you run `. ~/.config/sn-invoices/env` in your shell.

```
export SN_EMAIL='you@example.com'
export SN_PASS='your-password'
export SN_URL='https://acme.service-now.com/cts?id=invoices'
export SN_OUT="$HOME/Drive/invoices"
export SN_NAME='acme'
```

4. Create `local.mk` in the repo root with your portal URL, the output directory and the firm name. Git ignores this file.

```
.PHONY: acme

acme: build
	. "$$HOME/.config/sn-invoices/env" && ./sn-invoices --url 'https://acme.service-now.com/cts?id=invoices' --out "$$HOME/Drive/invoices" --name acme
```

5. Run it.

```
make acme
```

## Credentials

- Without `--headed`, the program reads `SN_EMAIL` and `SN_PASS`. If they are unset, it asks for them in the terminal.
- The login has two steps: the email, then the password. A portal with MFA is not handled.
- With `--headed`, log in by hand in the browser window. The program waits up to 5 minutes for the invoices.

## Schedule

`make install-timer` installs a systemd user timer that runs `dotenvx run -- scripts/scheduled-run.sh` on the 1st, 11th and 21st at 08:00, and after boot when a run was missed.

The variables come from the dotenvx `.env` file in the repo, which needs `SN_EMAIL`, `SN_PASS`, `SN_URL`, `SN_OUT`, `SN_NAME` and `NTFY_TOPIC`, a secret topic name. Subscribe to that topic in the ntfy app to get a push notification when `make run` fails, or on the 21st or later when `SN_OUT` has no file from the last 31 days.

## Output

Files are named `<yyyy-mm-dd>_<NAME>_<original name>`, and existing files are skipped. `NAME` defaults to the first label of the URL host.

## Troubleshooting

When no invoice rows are found, the program saves the current page to `debug-page.html` in the working directory. That file contains your invoice data, so do not commit or share it.

## Development

`make verify` runs the prerequisite check, gofmt, vet, build and tests.
