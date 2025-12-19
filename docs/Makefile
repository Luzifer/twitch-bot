DOCS_BASE_URL:=/
HUGO_VERSION:=0.152.2

HUGO := ./hugo_$(HUGO_VERSION)

render_docs: $(HUGO) ## Render documentation site
	$(HUGO) \
		--baseURL "$(DOCS_BASE_URL)" \
		--cleanDestinationDir \
		--gc

$(HUGO):
	curl -sSfL https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_linux-amd64.tar.gz | tar -xz hugo
	mv hugo $@
