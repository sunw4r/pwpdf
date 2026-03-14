WAILS ?= wails
WAILS_TAGS ?= webkit2_41

.PHONY: dev build build-all test frontend-install frontend-build cli

dev:
	$(WAILS) dev -tags $(WAILS_TAGS)

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

test:
	go test ./...

build:
	go run ./scripts/syncicons
	$(MAKE) test
	$(WAILS) build -clean -tags $(WAILS_TAGS)

build-all:
	@test -n "$(SELECTOR)" || (echo "SELECTOR is required. Example: make build-all SELECTOR=all" >&2; exit 1)
	./scripts/build-all.sh --selector "$(SELECTOR)"

cli:
	go build -o build/bin/pwpdf .
