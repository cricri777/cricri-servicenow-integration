-include local.mk

URL ?=
OUT ?=
NAME ?=

.PHONY: check build verify run

check:
	./scripts/check-prereqs.sh

build:
	go build -o sn-invoices .

verify: check
	./scripts/verify.sh

run: build
	./scripts/check-prereqs.sh
	./sn-invoices --url "$(URL)" --out "$(OUT)" --name "$(NAME)"
