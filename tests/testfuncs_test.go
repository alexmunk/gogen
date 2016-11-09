package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSampleInFile(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := filepath.Join("..", "tests", "tokens")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := FindSampleInFile(home, "token-static")
	if s == nil {
		t.Fatalf("Sample token-static not found in file: %s", home)
	}
}
