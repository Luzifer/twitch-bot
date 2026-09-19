## Tool Binaries
GO_RUN := go run -modfile ./tools/go.mod
GO_TEST = $(GO_RUN) gotest.tools/gotestsum --format pkgname
GOLANCI_LINT = golangci-lint

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

build_prod: frontend_prod overlays ## Build release binary locally
	go build \
		-trimpath \
		-mod=readonly \
		-ldflags "-X main.version=$(shell git describe --tags --always || echo dev)"

bundle_eventclient: ## Creates an NPM package containing eventclient and types
bundle_eventclient: reset_builddir overlays
	install -Dm0644 -t ci/npm-bundle/dist \
		internal/apimodules/overlays/default/eventclient.d.ts \
		internal/apimodules/overlays/default/eventclient.js \
		internal/apimodules/overlays/default/eventTypes.d.ts \
		internal/apimodules/overlays/default/eventTypes.js
	pnpm pack --dir $(CURDIR)/ci/npm-bundle --out $(CURDIR)/.build/%s-%v.tgz

publish: ## Run build tooling to produce all binaries
publish: reset_builddir frontend_prod overlays bundle_eventclient
	bash ./ci/build.sh

reset_builddir:
	rm -rf $(CURDIR)/.build
	install -dm0755 $(CURDIR)/.build

##@ Development

lint: ## Run Linter against code
	$(GOLANCI_LINT) run ./...

short_test: ## Run tests not depending on network
	$(GO_TEST) --hide-summary skipped -- ./... -cover -short

test: ## Run all tests
	$(GO_TEST) --hide-summary skipped -- ./... -cover

##@ Editor frontend

frontend_prod: export NODE_ENV=production
frontend_prod: frontend ## Build frontend in production mode

frontend: node_modules ## Build frontend
	pnpm node ci/build.mjs

frontend_lint: node_modules ## Lint frontend files
	pnpm eslint \
		--fix \
		src

overlays: node_modules ## Build default overlays
	$(MAKE) -C internal/apimodules/overlays build

node_modules: ## Install node modules
	pnpm i --frozen-lockfile

##@ Tooling

update-chrome-major: ## Patch latest Chrome major version into linkcheck
	sed -i -E \
		's/chromeMajor = [0-9]+/chromeMajor = $(shell curl -sSf https://lv.luzifer.io/v1/catalog/google-chrome/stable/version | cut -d '.' -f 1)/' \
		internal/linkcheck/useragent.go

##@ Vulnerability scanning

trivy: ## Run Trivy against the code
	trivy fs . \
		--dependency-tree \
		--exit-code 1 \
		--format table \
		--ignore-unfixed \
		--quiet \
		--scanners misconfig,license,secret,vuln \
		--severity HIGH,CRITICAL \
		--skip-dirs docs,tools

##@ Documentation

docs: generate_docs ## Generate all documentation

generate_docs: ## Generate project documentation
	go run -tags docgen . --storage-conn-string $(shell mktemp --suffix=.db) generate-docs

render_docs: ## Render documentation site
	$(MAKE) -C docs
