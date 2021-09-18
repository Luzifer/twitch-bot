package plugins

type (
	ActionDocumentation struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Type        string `json:"type"`

		Fields []ActionDocumentationField `json:"fields"`
	}

	ActionDocumentationField struct {
		Default         string                       `json:"default"`
		Description     string                       `json:"description"`
		Key             string                       `json:"key"`
		Name            string                       `json:"name"`
		Optional        bool                         `json:"optional"`
		SupportTemplate bool                         `json:"support_template"`
		Type            ActionDocumentationFieldType `json:"type"`
	}

	ActionDocumentationFieldType string
)

const (
	ActionDocumentationFieldTypeBool        ActionDocumentationFieldType = "bool"
	ActionDocumentationFieldTypeDuration    ActionDocumentationFieldType = "duration"
	ActionDocumentationFieldTypeInt64       ActionDocumentationFieldType = "int64"
	ActionDocumentationFieldTypeString      ActionDocumentationFieldType = "string"
	ActionDocumentationFieldTypeStringSlice ActionDocumentationFieldType = "stringslice"
)
