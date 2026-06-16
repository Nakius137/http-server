package web

import "embed"

//go:embed templates/layouts/*.html templates/pages/*.html
var TemplatesFS embed.FS
