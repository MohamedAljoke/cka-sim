package catalog

import "embed"

//go:embed lib.sh */task.md */explain.md */setup.sh */check.sh
var FS embed.FS
