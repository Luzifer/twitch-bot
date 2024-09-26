DOCS_BASE_URL:=/
HUGO_VERSION:=0.117.0

## Tool Binaries
GO_RUN := go run -modfile ./tools/go.mod
GO_TEST = $(GO_RUN) gotest.tools/gotestsum --format pkgname
GOLANCI_LINT = $(GO_RUN) github.com/golangci/golangci-lint/cmd/golangci-lint

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

build_prod: frontend_prod ## Build release binary locally
	go build \
		-trimpath \
		-mod=readonly \
		-ldflags "-X main.version=$(shell git describe --tags --always || echo dev)"

publish: frontend_prod ## Run build tooling to produce all binaries
	bash ./ci/build.sh

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
	node ci/build.mjs

frontend_lint: node_modules ## Lint frontend files
	./node_modules/.bin/eslint \
		--ext .js,.vue \
		--fix \
		src

node_modules: ## Install node modules
	npm ci --include dev

##@ Tooling

update-chrome-major: ## Patch latest Chrome major version into linkcheck
	sed -i -E \
		's/chromeMajor = [0-9]+/chromeMajor = $(shell curl -sSf https://lv.luzifer.io/v1/catalog/google-chrome/stable/version | cut -d '.' -f 1)/' \
		internal/linkcheck/useragent.go

gh-workflow: ## Regenerate CI workflow
	bash ci/create-workflow.sh

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

docs: actor_docs eventclient_docs template_docs ## Generate all documentation

actor_docs: ## Generate actor documentation
	go run . --storage-conn-string $(shell mktemp).db actor-docs >docs/content/configuration/actors.md

template_docs: ## Generate template function documentation
	go run . --storage-conn-string $(shell mktemp).db tpl-docs >docs/content/configuration/templating.md

eventclient_docs: ## Generate eventclient documentation
	echo -e "---\ntitle: EventClient\nweight: 10000\n---\n" >docs/content/overlays/eventclient.md
	docker run --rm -i -v $(CURDIR):$(CURDIR) -w $(CURDIR) node:18-alpine sh -ec 'npx --yes jsdoc-to-markdown --files ./internal/apimodules/overlays/default/eventclient.js' >>docs/content/overlays/eventclient.md

render_docs: hugo_$(HUGO_VERSION) ## Render documentation site
	./hugo_$(HUGO_VERSION) \
		--baseURL "$(DOCS_BASE_URL)" \
		--cleanDestinationDir \
		--gc \
		--source docs

hugo_$(HUGO_VERSION):
	curl -sSfL https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_linux-amd64.tar.gz | tar -xz hugo
	mv hugo $@
