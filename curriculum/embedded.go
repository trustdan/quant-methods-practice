package curriculum

import "embed"

// ApprovedFS embeds all approved curriculum question template JSON files.
//
//go:embed approved/*.json
var ApprovedFS embed.FS
