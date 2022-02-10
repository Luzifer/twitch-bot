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
	go run . --storage-file $(shell mktemp).json.gz actor-docs >wiki/Actors.md

pull_wiki:
	git subtree pull --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master --squash

push_wiki:
	git subtree push --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master
