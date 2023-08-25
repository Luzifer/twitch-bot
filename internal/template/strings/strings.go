package strings

import (
	"encoding/base64"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterTemplateFunction("b64urlenc", plugins.GenericTemplateFunctionGetter(base64URLEncode), plugins.TemplateFuncDocumentation{
		Description: "Encodes the input using base64 URL-encoding (like `b64enc` but using `URLEncoding` instead of `StdEncoding`)",
		Syntax:      "b64urlenc <input>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ b64urlenc "mystring" }}`,
			ExpectedOutput: "bXlzdHJpbmc=",
		},
	})

	args.RegisterTemplateFunction("b64urldec", plugins.GenericTemplateFunctionGetter(base64URLDecode), plugins.TemplateFuncDocumentation{
		Description: "Decodes the input using base64 URL-encoding (like `b64dec` but using `URLEncoding` instead of `StdEncoding`)",
		Syntax:      "b64urldec <input>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ b64urldec "bXlzdHJpbmc=" }}`,
			ExpectedOutput: "mystring",
		},
	})

	return nil
}

func base64URLEncode(v string) string {
	return base64.URLEncoding.EncodeToString([]byte(v))
}

func base64URLDecode(v string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(v)
	if err != nil {
		return "", errors.Wrap(err, "decoding string")
	}
	return string(data), nil
}
