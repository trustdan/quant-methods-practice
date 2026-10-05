package assets

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed dist/*
var embeddedFS embed.FS

// FS returns an fs.FS rooted at the compiled dist directory.
func FS() (fs.FS, error) {
	sub, err := fs.Sub(embeddedFS, "dist")
	if err != nil {
		return nil, fmt.Errorf("failed to open embedded dist filesystem: %w", err)
	}
	return sub, nil
}
