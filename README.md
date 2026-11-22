# cricri-servicenow-integration

Downloads the invoices from a ServiceNow portal into a local directory.

## Prerequisites

Go and Chrome or Chromium. Run `make check` to verify them.

## Usage

```
make run URL='https://fx.service-now.com/cts?id=invoices' OUT=~/Drive/invoices NAME=fx
```

The program asks for the email and the password when `SN_EMAIL` and `SN_PASS` are not set.

Files are named `<yyyy-mm-dd>_<NAME>_<original name>`. Existing files are skipped. `NAME` defaults to the first label of the URL host.
