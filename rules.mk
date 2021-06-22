SUBDIRS := $(wildcard */*/Makefile */Makefile)
BUILD_DIRS := $(addsuffix /build,$(shell find ./cmd/* -type d 2>/dev/null))
INSTALL_DIRS := $(addsuffix /install,$(shell find ./cmd/* -type d 2>/dev/null))
LINT_DIRS := $(subst Makefile,lint,$(SUBDIRS))
COVER_DIRS := $(subst Makefile,cover,$(SUBDIRS))
DOC_DIRS := $(subst Makefile,doc,$(SUBDIRS))
STATICCHECK = $(HOME)/go/bin/staticcheck
GO_FILES = $(wildcard *.go)
DOCKER_APP = docker_$(APP)
DOCKER_IMAGE = $(APP).docker
GOBIN := $(shell while true; do if [[ `pwd` == "/" ]]; then exit 1;fi;if [[ -f `pwd`/go.mod ]]; then echo `pwd`/bin;exit 0;fi;cd ..;done)
export GOBIN

# When using private repos, run this command:
# git config -global url."ssh://git@github.com/mediamath".insteadOf "https://github.com/mediamath"
# or add the following lines to your ~/.gitconfig
# [url "ssh://git@github.com/mediamath"]
#   insteadOf = https://github.com/mediamath
#
export GOPRIVATE=github.com/MediaMath/*,github.com/mediamath/*

CODE_COVERAGE_PERCENT := $(if $(CODE_COVERAGE_PERCENT),$(CODE_COVERAGE_PERCENT),60.0)


.PHONY: all $(SUBDIRS) build $(INSTALL_DIRS)
all: build
ifneq ($(SUBDIRS),)
	@echo "+++ Building"  $(notdir $(CURDIR))
	go fmt ./...
	go vet ./...
	#go test -cover --tags unit ./...
endif

install: $(INSTALL_DIRS)
ifneq ($(GO_FILES),)
	@echo "+++ Building"  $(notdir $(CURDIR))
	go fmt
	go vet
	go install
endif
$(INSTALL_DIRS) $(BUILD_DIRS):
	@$(MAKE) -C $(@D) $(@F)

#
# plain make
#
build: $(BUILD_DIRS)
ifneq ($(GO_FILES),)
	@echo "+++ Building $(notdir $(CURDIR))"
	go fmt
	go vet
	go test -cover --tags unit
	go build
endif


lint: $(LINT_DIRS)
ifneq ($(GO_FILES),)
	go vet
endif
$(LINT_DIRS):
	$(MAKE) -C $(@D) lint

$(SUBDIRS): build
	@$(MAKE) -C $(@D)

$(STATICCHECK):
	go get honnef.co/go/tools/cmd/staticcheck

#
# different test modes, unit, integration, test (all tests)
#
.PHONY: unit
unit:
	go test -cover -race -coverprofile=coverage_report.txt -covermode=atomic --tags unit ./...

.PHONY: integration
integration:
	go test -cover --tags integration ./...

.PHONY: test
test:
	go test -cover --tags all ./... | awk '{print $0};/coverage:/ {if ($(CODE_COVERAGE_PERCENT) > $$2) {print "CODE COVERAGE < $(CODE_COVERAGE_PERCENT)%"; exit 1}}'

$(INTEGRATION_DIRS) $(TEST_DIRS) $(DOC_DIRS):
	$(MAKE) -C $(@D) $(@F)

.PHONY: showcover
showcover:
	go test --tags all -coverprofile=c.out && go tool cover -html=c.out

#
# test coverage report
#
cover: $(COVER_DIRS)
ifneq ($(GO_FILES),)
	@echo $$(basename $$(pwd)) "*****************" 
	-@go test -cover --tags all| grep -e coverage -e "no test";
endif
$(COVER_DIRS):
	@$(MAKE)  --no-print-directory -C $(@D) cover

.PHONY: doc
doc: $(DOC_DIRS)
ifneq ($(GO_FILES),)
	go doc -all > GoDOC.md
endif

.PHONY: docker
docker: $(DOCKER_IMAGE)
$(DOCKER_IMAGE):
	$$(cd ecs;go mod vendor)
	docker build .
	$$(cd ecs;rm -rf vendor)

.PHONY: docker_minion
docker_minion:
	go mod vendor
	docker build --file ./Dockerfile ../../..
	rm -rf ../../vendor

$(DOCKER_APP): all
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o $@ .
