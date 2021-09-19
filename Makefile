default: lint test

lint:
	golangci-lint run --timeout=5m

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh

test:
	go test -cover -v ./...

# --- Editor frontend

editor/bundle.js:
	bash ci/bundle.sh $@ \
		npm/axios@0.21.4/dist/axios.min.js \
		npm/vue@2 \
		npm/bootstrap-vue@2/dist/bootstrap-vue.min.js \
		npm/moment@2

editor/bundle.css:
	bash ci/bundle.sh $@ \
		npm/bootstrap@4/dist/css/bootstrap.min.css \
		npm/bootstrap-vue@2/dist/bootstrap-vue.min.css \
		npm/bootswatch@4/dist/darkly/bootstrap.min.css

# --- Wiki Updates

pull_wiki:
	git subtree pull --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master --squash

push_wiki:
	git subtree push --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master
