package assets_test

import (
	"io/fs"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/assets"
)

func TestFS(t *testing.T) {
	distFS, err := assets.FS()
	if err != nil {
		t.Fatalf("expected assets.FS() to succeed, got %v", err)
	}

	// Verify index.html exists in embedded filesystem
	info, err := fs.Stat(distFS, "index.html")
	if err != nil {
		t.Fatalf("expected index.html to exist in embedded filesystem: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("expected index.html to be a file, but it is a directory")
	}
	if info.Size() == 0 {
		t.Fatalf("expected index.html to have non-zero size")
	}
}
