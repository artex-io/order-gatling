PROJECT_NAME      ?= order-gatling
GO                ?= $(shell which go)
GH                ?= $(shell which gh)
GIT_UPDATE_INDEX  ?= $(shell git update-index --refresh)
GIT_REVISION      ?= $(shell git rev-parse HEAD)
GIT_VERSION       ?= $(shell git describe --tags --always --dirty --exclude "deploy-*" 2>/dev/null || echo dev)

GOENV_GOOS               := $(shell go env GOOS)
GOENV_GOARCH             := $(shell go env GOARCH)
GOENV_GOARM              := $(shell go env GOARM)
GOENV_GOPRIVATE          := $(shell go env GOPRIVATE)
GOOS                     ?= $(GOENV_GOOS)
GOARCH                   ?= $(GOENV_GOARCH)
GOARM                    ?= $(GOENV_GOARM)
GOPRIVATE                ?= $(GOENV_GOPRIVATE)

GO_BUILD_SRC             := $(shell find . -name \*.go -type f) go.mod go.sum
GO_BUILD_EXTLDFLAGS      :=
GO_BUILD_TAGS            :=
GO_BUILD_TARGET_DEPS     :=
GO_BUILD_FLAGS           := -trimpath
GO_BUILD_LDFLAGS_OPTIMS  ?=

ifeq ($(GOOS)/$(GOARCH),$(GOENV_GOOS)/$(GOENV_GOARCH))
GO_BUILD_TARGET          ?= dist/$(PROJECT_NAME)
GO_BUILD_VERSION_TARGET  ?= dist/$(PROJECT_NAME)-$(GIT_VERSION)
else
ifeq ($(GOARCH),arm)
GO_BUILD_TARGET          ?= dist/$(PROJECT_NAME)-$(GOOS)-$(GOARCH)v$(GOARM)
GO_BUILD_VERSION_TARGET  := dist/$(PROJECT_NAME)-$(GIT_VERSION)-$(GOOS)-$(GOARCH)v$(GOARM)
else
GO_BUILD_TARGET          ?= dist/$(PROJECT_NAME)-$(GOOS)-$(GOARCH)
GO_BUILD_VERSION_TARGET  := dist/$(PROJECT_NAME)-$(GIT_VERSION)-$(GOOS)-$(GOARCH)
endif # ($(GOARCH),arm)
endif # ($(GOOS)/$(GOARCH),$(GOENV_GOOS)/$(GOENV_GOARCH))

ifneq ($(DEBUG),)
GO_BUILD_FLAGS            = -gcflags="all=-N -l"
else
GO_BUILD_LDFLAGS_OPTIMS  += -s -w
endif # $(DEBUG)

GO_BUILD_FLAGS_TARGET           := .go-build-flags
GO_CROSSBUILD_WINDOWS_PLATFORMS := windows/amd64 windows/386 windows/arm
GO_CROSSBUILD_PLATFORMS         ?= linux/386 linux/amd64 linux/arm linux/arm64 \
                                   darwin/amd64 darwin/arm64

GO_CROSSBUILD_386_PLATFORMS           := $(filter %/386,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_AMD64_PLATFORMS         := $(filter %/amd64,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_ARM_PLATFORMS           := $(filter %/arm,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_ARM64_PLATFORMS         := $(filter %/arm64,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_WINDOWS_PLATFORMS       := $(filter windows/%,$(GO_CROSSBUILD_WINDOWS_PLATFORMS))

GO_CROSSBUILD_386_TARGET_PATTERN      := dist/$(PROJECT_NAME)-$(GIT_VERSION)-%-386
GO_CROSSBUILD_AMD64_TARGET_PATTERN    := dist/$(PROJECT_NAME)-$(GIT_VERSION)-%-amd64
GO_CROSSBUILD_ARM_TARGET_PATTERN      := dist/$(PROJECT_NAME)-$(GIT_VERSION)-%-arm
GO_CROSSBUILD_ARM64_TARGET_PATTERN    := dist/$(PROJECT_NAME)-$(GIT_VERSION)-%-arm64
GO_CROSSBUILD_WINDOWS_TARGET_PATTERN  := dist/$(PROJECT_NAME)-$(GIT_VERSION)-windows-%.exe

GO_CROSSBUILD_TARGETS := $(patsubst %/386,$(GO_CROSSBUILD_386_TARGET_PATTERN),$(GO_CROSSBUILD_386_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/amd64,$(GO_CROSSBUILD_AMD64_TARGET_PATTERN),$(GO_CROSSBUILD_AMD64_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/arm,$(GO_CROSSBUILD_ARM_TARGET_PATTERN),$(GO_CROSSBUILD_ARM_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/arm64,$(GO_CROSSBUILD_ARM64_TARGET_PATTERN),$(GO_CROSSBUILD_ARM64_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst windows/%,$(GO_CROSSBUILD_WINDOWS_TARGET_PATTERN),$(GO_CROSSBUILD_WINDOWS_PLATFORMS))

GO_BUILD_EXTLDFLAGS     := $(strip $(GO_BUILD_EXTLDFLAGS))
GO_BUILD_TAGS           := $(strip $(GO_BUILD_TAGS))
GO_BUILD_TARGET_DEPS    := $(strip $(GO_BUILD_TARGET_DEPS))
GO_BUILD_FLAGS          := $(strip $(GO_BUILD_FLAGS))
GO_BUILD_LDFLAGS_OPTIMS := $(strip $(GO_BUILD_LDFLAGS_OPTIMS))
GO_BUILD_LDFLAGS        := -ldflags '$(GO_BUILD_LDFLAGS_OPTIMS) -extldflags "$(GO_BUILD_EXTLDFLAGS)"'

GO_TOOLS_GOLANGCI_LINT ?= $(shell $(GO) env GOPATH)/bin/golangci-lint

# ------------------------------------------------------------------------------

.PHONY: all build crossbuild crossbuild-checksums build-deps .FORCE

all: crossbuild crossbuild-checksums

build: $(GO_BUILD_VERSION_TARGET) $(GO_BUILD_TARGET)

install:
	CGO_ENABLED=0 $(GO) install -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS)

$(GO_BUILD_FLAGS_TARGET) : .FORCE
	@(echo "GO_VERSION=$(shell $(GO) version)"; \
	  echo "GO_GOOS=$(GOOS)"; \
	  echo "GO_GOARCH=$(GOARCH)"; \
	  echo "GO_GOARM=$(GOARM)"; \
	  echo "GO_BUILD_TAGS=$(GO_BUILD_TAGS)"; \
	  echo "GO_BUILD_FLAGS=$(GO_BUILD_FLAGS)"; \
	  echo 'GO_BUILD_LDFLAGS=$(subst ','\'',$(GO_BUILD_LDFLAGS))') \
	    | cmp -s - $@ \
	        || (echo "GO_VERSION=$(shell $(GO) version)"; \
	            echo "GO_GOOS=$(GOOS)"; \
	            echo "GO_GOARCH=$(GOARCH)"; \
	            echo "GO_GOARM=$(GOARM)"; \
	            echo "GO_BUILD_TAGS=$(GO_BUILD_TAGS)"; \
	            echo "GO_BUILD_FLAGS=$(GO_BUILD_FLAGS)"; \
	            echo 'GO_BUILD_LDFLAGS=$(subst ','\'',$(GO_BUILD_LDFLAGS))') > $@

$(GO_BUILD_TARGET): $(GO_BUILD_VERSION_TARGET)
	@(test -e $@ && unlink $@) || true
	@mkdir -p $$(dirname $@)
	@ln $< $@

build-deps:
	for dep in $$(grep -E "(sylr/quickfixgo|quickfixgo)" go.mod | awk '{ print $$1; }'); do \
	  CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) $$dep; \
	done

$(GO_BUILD_VERSION_TARGET): $(GO_BUILD_SRC) $(GO_GENERATE_TARGET) $(GO_BUILD_FLAGS_TARGET) | $(GO_BUILD_TARGET_DEPS)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

crossbuild: $(GO_BUILD_VERSION_TARGET) $(GO_CROSSBUILD_TARGETS)

$(GO_CROSSBUILD_WINDOWS_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=windows GOARCH=$* $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_386_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=386 $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_AMD64_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=amd64 $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_ARM_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=arm $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_ARM64_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=arm64 $(GO) build -tags "$(GO_BUILD_TAGS)" $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

crossbuild-checksums: dist/checksums

dist/checksums : $(GO_CROSSBUILD_TARGETS)
	cd dist && shasum -a 256 $(PROJECT_NAME)-*-* > checksums

# -- go mod --------------------------------------------------------------------

.PHONY: go-mod-verify go-mod-tidy

go-mod-verify:
	$(GO) mod download
	git diff --quiet go.* || git diff --exit-code go.* || exit 1

go-mod-tidy:
	$(GO) mod download
	$(GO) mod tidy

# ------------------------------------------------------------------------------

test:
	go test ./...

lint: $(GO_TOOLS_GOLANGCI_LINT)
	$(GO_TOOLS_GOLANGCI_LINT) run

# -- tools ---------------------------------------------------------------------

.PHONY: tools

tools: $(GO_TOOLS_GOLANGCI_LINT)

$(GO_TOOLS_GOLANGCI_LINT):
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
