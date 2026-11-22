# cricri-servicenow-integration

Downloads the invoices from a ServiceNow portal into a local directory.

This tool is custom for one client portal and one firm. It reads the invoice rows of that portal page and clicks each row, so it will not work as is on other ServiceNow portals.

## Prerequisites

Go and Chrome or Chromium. Run `make check` to verify them.

## Usage

```
make run URL='https://fx.service-now.com/cts?id=invoices' OUT=~/Drive/invoices NAME=fx
```

With `--headed`, log in by hand in the browser window and the program waits up to 5 minutes for the invoices. Otherwise the program asks for the email and the password when `SN_EMAIL` and `SN_PASS` are not set.

Files are named `<yyyy-mm-dd>_<NAME>_<original name>`. Existing files are skipped. `NAME` defaults to the first label of the URL host.
