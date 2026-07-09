BINARY      := wolproxy
PREFIX      ?= /usr/local
BINDIR      := $(PREFIX)/bin
SYSTEMD_DIR ?= /etc/systemd/system
ENV_FILE    ?= /etc/wolproxy.env

# Cross-compilation target for `build-linux` (override GOARCH=arm for 32-bit ARM Linux).
GOARCH ?= arm64

.PHONY: all build build-linux vet tidy clean install uninstall

all: build

build:
	go build -o $(BINARY) .

build-linux:
	GOOS=linux GOARCH=$(GOARCH) go build -o $(BINARY) .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)

install: build
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -Dm644 wolproxy.service $(DESTDIR)$(SYSTEMD_DIR)/wolproxy.service
	test -f $(DESTDIR)$(ENV_FILE) || install -Dm600 wolproxy.env $(DESTDIR)$(ENV_FILE)
	@echo "Installed. Then: systemctl daemon-reload && systemctl enable --now wolproxy"

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(SYSTEMD_DIR)/wolproxy.service
