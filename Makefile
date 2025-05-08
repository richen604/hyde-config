# Makefile for hyde-config

GO = go
NAME = hyde-config
PREFIX = $(HOME)/.local
LIBDIR = $(PREFIX)/lib/hyde
CFGDIR = $(HOME)/.config/hyde
STATEDIR = $(HOME)/.local/state/hyde
SVCDIR = $(HOME)/.config/systemd/user
BUILDDIR = build
RELEASEDIR = bin

.PHONY: all build install uninstall clean run service-install service-enable service-disable

all: build

build:
	mkdir -p $(BUILDDIR)
	$(GO) build -ldflags="-s -w" -o $(BUILDDIR)/$(NAME)

release:
	mkdir -p $(RELEASEDIR)
	$(GO) build -ldflags="-s -w" -o $(RELEASEDIR)/$(NAME)

install: build
	mkdir -p $(LIBDIR) $(CFGDIR) $(STATEDIR)
	install -m755 $(BUILDDIR)/$(NAME) $(LIBDIR)/$(NAME)
	@echo "Installed to $(LIBDIR)/$(NAME)"

service-install:
	mkdir -p $(SVCDIR)
	install -m644 hyde-config.service $(SVCDIR)/hyde-config.service
	@echo "Service installed"

service-enable:
	systemctl --user daemon-reload
	systemctl --user enable hyde-config.service
	@echo "Service enabled"
	
service-start:
	systemctl --user start hyde-config.service
	@echo "Service started"

service-stop:
	systemctl --user stop hyde-config.service
	@echo "Service stopped"

service-disable:
	systemctl --user disable hyde-config.service
	@echo "Service disabled"
	
uninstall:
	rm -f $(LIBDIR)/$(NAME)
	rm -f $(SVCDIR)/hyde-config.service
	@echo "Uninstalled"

clean:
	rm -rf $(BUILDDIR) $(RELEASEDIR)
	$(GO) clean