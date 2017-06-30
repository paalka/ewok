GO_EXECUTABLE ?= go
PACKAGE_DIRS := $(shell glide nv)
BINDIR := $(CURDIR)/bin

.PHONY: build
build:
	GOBIN=$(BINDIR) ${GO_EXECUTABLE} install github.com/paalka/ewok/cmd/...

.PHONY: deps
deps:
	glide install --strip-vendor
	mkdir -p $(BINDIR)
