package storage

import (
	"os"
	"path/filepath"
	"runtime"
)

// AppName is the standard application identifier for user data directories.
const AppName = "quant-methods-practice"

// DefaultDataDir returns the platform-specific user data directory for quant-methods-practice.
// Windows: APPDATA/quant-methods-practice (or AppData/Roaming/quant-methods-practice)
// macOS: Library/Application Support/quant-methods-practice
// Linux/Unix: XDG_DATA_HOME/quant-methods-practice or ~/.local/share/quant-methods-practice
func DefaultDataDir() (string, error) {
	if envDir := os.Getenv("QUANT_DATA_DIR"); envDir != "" {
		return envDir, nil
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, AppName), nil

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", AppName), nil

	default: // Linux, Unix, etc.
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, AppName), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", AppName), nil
	}
}

// ResolveDBPath resolves the final SQLite database path from flags or default data dir.
func ResolveDBPath(dataDirFlag, dbFlag string) (string, error) {
	if dbFlag != "" {
		return dbFlag, nil
	}

	dataDir := dataDirFlag
	if dataDir == "" {
		var err error
		dataDir, err = DefaultDataDir()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(dataDir, "quant-methods.db"), nil
}
