# go option
GO         ?= go
PKG        := go mod vendor
LDFLAGS    := -w -s
GOFLAGS    :=
TAGS       := 
BINDIR     := $(CURDIR)/bin
PKGDIR     := github.com/tritonmedia/twilight.go

# Required for globs to work correctly
SHELL=/bin/bash


.PHONY: all
all: build

.PHONY: dep
dep:
	@echo " ===> Installing dependencies via '$$(awk '{ print $$1 }' <<< "$(PKG)" )' <=== "
	@$(PKG)

.PHONY: build
build:
	@echo " ===> building releases in ./bin/... <=== "
	GO11MODULE=enabled CGO_ENABLED=1 $(GO) build -o $(BINDIR)/twilight -v $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' .

.PHONY: gofmt
gofmt:
	@echo " ===> Running go fmt <==="
	gofmt -w ./

.PHONY: test
test:
	go test $(PKGDIR)/...