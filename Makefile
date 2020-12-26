default:

pull_wiki:
	git subtree pull --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master --squash

push_wiki:
	git subtree push --prefix=wiki https://github.com/Luzifer/twitch-bot.wiki.git master
