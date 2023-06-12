default: lint frontend_lint test

lint:
	golangci-lint run

publish: frontend
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh

test:
	go test -cover -v ./...

# --- Editor frontend

frontend: node_modules
	node ci/build.mjs

frontend_lint: node_modules
	./node_modules/.bin/eslint \
		--ext .js,.vue \
		--fix \
		src

node_modules:
	npm ci

# --- Wiki Updates

actor_docs:
	go run . --storage-conn-string $(shell mktemp).db actor-docs >wiki/Actors.md

pull_wiki:
	git subtree pull --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master --squash

push_wiki:
	git subtree push --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master

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
