SUBDIRS := $(wildcard */*/Makefile */Makefile)
BUILD_DIRS := $(addsuffix /build,$(shell find ./cmd/* -type d 2>/dev/null))
GO_FILES = $(wildcard *.go)

.PHONY: all $(SUBDIRS) build

all: build
ifneq ($(SUBDIRS),)
    @echo "+++ Building"  $(notdir $(CURDIR))
    go fmt ./...
    go vet ./...
    go test -cover --tags unit ./...
endif

build: $(BUILD_DIRS)
ifneq ($(GO_FILES),)
    @echo "+++ Building $(notdir $(CURDIR))"
    go fmt
    go vet
    go test -cover --tags unit
    go build
endif

.PHONY: showcover
showcover:
    go test --tags all -coverprofile=c.out && go tool cover -html=c.out
