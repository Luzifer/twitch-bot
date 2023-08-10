HUGO_VERSION:=0.117.0

default: lint frontend_lint test

lint:
	golangci-lint run

publish: frontend_prod
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh

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
	npm ci

# --- Tools

update_ua_list:
	# User-Agents provided by https://www.useragents.me/
	curl -sSf https://www.useragents.me/api | jq -r '.data[].ua' | grep -v 'Trident' >internal/linkcheck/user-agents.txt

# -- Vulnerability scanning --

trivy:
	trivy fs . \
		--dependency-tree \
		--exit-code 1 \
		--format table \
		--ignore-unfixed \
		--quiet \
		--scanners config,license,secret,vuln \
		--severity HIGH,CRITICAL \
		--skip-dirs docs

# -- Documentation Site --

actor_docs:
	go run . --storage-conn-string $(shell mktemp).db actor-docs >docs/content/configuration/actors.md

render_docs: hugo_$(HUGO_VERSION)
	./hugo_$(HUGO_VERSION) --cleanDestinationDir --gc --source docs

hugo_$(HUGO_VERSION):
	curl -sSfL https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_linux-amd64.tar.gz | tar -xz hugo
	mv hugo $@
