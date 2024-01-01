package plugins

type (
	// TemplateFuncDocumentation contains a documentation for a template
	// function to be rendered into the documentation site
	TemplateFuncDocumentation struct {
		Name        string
		Description string
		Syntax      string
		Example     *TemplateFuncDocumentationExample
		Remarks     string
	}

	// TemplateFuncDocumentationExample contains an example of the
	// function execution to be rendered as an example how to use the
	// template function
	TemplateFuncDocumentationExample struct {
		MatchMessage   string
		MessageContent string
		Template       string
		ExpectedOutput string
		FakedOutput    string
	}
)
