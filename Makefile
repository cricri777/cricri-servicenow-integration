-include local.mk

URL ?= $(SN_URL)
OUT ?= $(SN_OUT)
NAME ?= $(SN_NAME)

.PHONY: check build verify run install-timer

check:
	./scripts/check-prereqs.sh

build:
	go build -o sn-invoices .

verify: check
	./scripts/verify.sh

run: build
	./scripts/check-prereqs.sh
	./sn-invoices --url "$(URL)" --out "$(OUT)" --name "$(NAME)"

install-timer:
	mkdir -p "$$HOME/.config/systemd/user"
	sed "s|@DIR@|$(CURDIR)|g" systemd/sn-invoices.service > "$$HOME/.config/systemd/user/sn-invoices.service"
	cp systemd/sn-invoices.timer "$$HOME/.config/systemd/user/sn-invoices.timer"
	systemctl --user daemon-reload
	systemctl --user enable --now sn-invoices.timer
