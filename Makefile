VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
VERBOSE_FLAG = $(if $(VERBOSE),-v)
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u} $(VERBOSE_FLAG)
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/Songmu/godzil/cmd/godzil@latest
	go install github.com/tcnksm/ghr@latest

.PHONY: test
test: deps
	go test $(VERBOSE_FLAG) ./...

.PHONY: lint
lint: devel-deps
	# nop for now
	#   staticcheck fails with following error:
	#   ast.Package is deprecated: use the type checker [go/types] instead;
	# staticcheck -checks all ./...

.PHONY: build
build: deps
	go build $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS) ./cmd/gobump

.PHONY: install
install: deps
	go install $(VERBOSE_FLAG) -ldflags=$(BUILD_LDFLAGS) ./cmd/gobump

.PHONY: release
release: devel-deps
	godzil release

CREDITS: devel-deps go.sum
	godzil credits -w

DIST_DIR = dist/v$(VERSION)
.PHONY: crossbuild
crossbuild: CREDITS
	rm -rf $(DIST_DIR)
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) -d $(DIST_DIR) ./cmd/*

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) $(DIST_DIR)
