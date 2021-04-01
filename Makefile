default:

lint:
	golangci-lint run

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh

# --- Wiki Updates

pull_wiki:
	git subtree pull --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master --squash

push_wiki:
	git subtree push --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master
