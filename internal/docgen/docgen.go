// Package docgen contains generators for project documentation
package docgen

import (
	"bytes"
	"fmt"
	"os"
)

const (
	actorDocsPath         = "docs/content/configuration/actors.md"
	documentationFileMode = 0o644
	eventTypesPath        = "internal/apimodules/overlays/src/eventTypes.ts"
	eventDocsPath         = "docs/content/configuration/events.md"
	tplDocsPath           = "docs/content/configuration/templating.md"
)

func writeDocument(path string, content []byte) error {
	if err := os.WriteFile(path, append(bytes.TrimSpace(content), '\n'), documentationFileMode); err != nil { //#nosec:G306 // Generated documentation is intended to be world-readable
		return fmt.Errorf("writing documentation to %q: %w", path, err)
	}

	return nil
}
