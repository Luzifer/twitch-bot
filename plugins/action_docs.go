package plugins

type (
	// ActionDocumentation contains the documentation for a single actor
	// to be rendered into the documentation site
	ActionDocumentation struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Type        string `json:"type"`

		Fields []ActionDocumentationField `json:"fields"`
	}

	// ActionDocumentationField documents fields available in the actor
	ActionDocumentationField struct {
		Default         string                       `json:"default"`
		DefaultComment  string                       `json:"default_comment"`
		Description     string                       `json:"description"`
		Key             string                       `json:"key"`
		Long            bool                         `json:"long"`
		Name            string                       `json:"name"`
		Optional        bool                         `json:"optional"`
		SupportTemplate bool                         `json:"support_template"`
		Type            ActionDocumentationFieldType `json:"type"`
	}

	// ActionDocumentationFieldType defines known field types
	ActionDocumentationFieldType string
)

// Enum of available field types
const (
	ActionDocumentationFieldTypeBool        ActionDocumentationFieldType = "bool"
	ActionDocumentationFieldTypeDuration    ActionDocumentationFieldType = "duration"
	ActionDocumentationFieldTypeInt64       ActionDocumentationFieldType = "int64"
	ActionDocumentationFieldTypeString      ActionDocumentationFieldType = "string"
	ActionDocumentationFieldTypeStringSlice ActionDocumentationFieldType = "stringslice"
)
