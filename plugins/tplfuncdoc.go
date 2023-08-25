package plugins

type (
	TemplateFuncDocumentation struct {
		Name        string
		Description string
		Syntax      string
		Example     *TemplateFuncDocumentationExample
		Remarks     string
	}

	TemplateFuncDocumentationExample struct {
		MatchMessage   string
		MessageContent string
		Template       string
		ExpectedOutput string
		FakedOutput    string
	}
)
