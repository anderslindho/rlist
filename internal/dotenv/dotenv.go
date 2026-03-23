package dotenv

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadFromAncestors loads the first .env found walking up from the working directory.
// Variables already set in the process environment are not overridden.
func LoadFromAncestors() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		p := filepath.Join(dir, ".env")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			_ = godotenv.Load(p)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}
