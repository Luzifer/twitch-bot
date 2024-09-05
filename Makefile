DOCS_BASE_URL:=/
HUGO_VERSION:=0.117.0

default: lint frontend_lint test

build_prod: frontend_prod
	go build \
		-trimpath \
		-mod=readonly \
		-ldflags "-X main.version=$(shell git describe --tags --always || echo dev)"

lint:
	golangci-lint run

publish: frontend_prod
	bash ./ci/build.sh

short_test:
	go test -cover -test.short -v ./...

test:
	go test -cover -v ./...

# --- Editor frontend

frontend_prod: export NODE_ENV=production
frontend_prod: frontend

frontend: node_modules
	node ci/build.mjs

frontend_lint: node_modules
	./node_modules/.bin/eslint \
		--ext .js,.vue \
		--fix \
		src

node_modules:
	npm ci --include dev

# --- Tools

update-chrome-major:
	sed -i -E \
		's/chromeMajor = [0-9]+/chromeMajor = $(shell curl -sSf https://lv.luzifer.io/v1/catalog/google-chrome/stable/version | cut -d '.' -f 1)/' \
		internal/linkcheck/useragent.go

gh-workflow:
	bash ci/create-workflow.sh

# -- Vulnerability scanning --

trivy:
	trivy fs . \
		--dependency-tree \
		--exit-code 1 \
		--format table \
		--ignore-unfixed \
		--quiet \
		--scanners misconfig,license,secret,vuln \
		--severity HIGH,CRITICAL \
		--skip-dirs docs

# -- Documentation Site --

docs: actor_docs eventclient_docs template_docs

actor_docs:
	go run . --storage-conn-string $(shell mktemp).db actor-docs >docs/content/configuration/actors.md

template_docs:
	go run . --storage-conn-string $(shell mktemp).db tpl-docs >docs/content/configuration/templating.md

eventclient_docs:
	echo -e "---\ntitle: EventClient\nweight: 10000\n---\n" >docs/content/overlays/eventclient.md
	docker run --rm -i -v $(CURDIR):$(CURDIR) -w $(CURDIR) node:18-alpine sh -ec 'npx --yes jsdoc-to-markdown --files ./internal/apimodules/overlays/default/eventclient.js' >>docs/content/overlays/eventclient.md

render_docs: hugo_$(HUGO_VERSION)
	./hugo_$(HUGO_VERSION) \
		--baseURL "$(DOCS_BASE_URL)" \
		--cleanDestinationDir \
		--gc \
		--source docs

hugo_$(HUGO_VERSION):
	curl -sSfL https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_linux-amd64.tar.gz | tar -xz hugo
	mv hugo $@
