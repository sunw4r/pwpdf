WAILS ?= wails
WAILS_TAGS ?= $(shell if [ "$$(uname -s)" = "Linux" ] && command -v pkg-config >/dev/null 2>&1 && pkg-config --exists webkit2gtk-4.1; then printf '%s' webkit2_41; fi)
WAILS_TAG_ARGS := $(if $(strip $(WAILS_TAGS)),-tags $(WAILS_TAGS),)

.PHONY: dev build build-all test frontend-install frontend-build cli

dev:
	$(WAILS) dev $(WAILS_TAG_ARGS)

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

test:
	go test ./...

build:
	./scripts/build.sh

build-all:
	@test -n "$(SELECTOR)" || (echo "SELECTOR is required. Example: make build-all SELECTOR=all" >&2; exit 1)
	./scripts/build-all.sh --selector "$(SELECTOR)"

cli:
	go build -o build/bin/pwpdf .
