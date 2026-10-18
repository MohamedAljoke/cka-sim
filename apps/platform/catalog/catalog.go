package catalog

import "embed"

//go:embed lib.sh */task.md */explain.md */setup.sh */check.sh */solution.sh */reset.sh
var FS embed.FS
