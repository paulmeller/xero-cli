package cmdutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func readInputFromFile(t *testing.T, content string) error {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "input.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("file", "", "")
	if err := cmd.Flags().Set("file", path); err != nil {
		t.Fatal(err)
	}

	_, err := ReadInput(cmd)
	return err
}

// Regression for the "JSON scalar input accepted" bug: ReadInput only checked json.Valid, which
// accepts any JSON value including bare scalars. A scalar silently wraps into something like
// {"Accounts":[null]} downstream instead of failing with a clear local error.
func TestReadInput_RejectsScalars(t *testing.T) {
	for _, in := range []string{"null", "true", "false", "42", `"just a string"`} {
		t.Run(in, func(t *testing.T) {
			err := readInputFromFile(t, in)
			if err == nil {
				t.Fatalf("ReadInput(%s) should have been rejected, got nil error", in)
			}
			if !strings.Contains(err.Error(), "object or array") {
				t.Errorf("error should explain object/array requirement, got: %v", err)
			}
		})
	}
}

func TestReadInput_AcceptsObject(t *testing.T) {
	if err := readInputFromFile(t, `{"Code":"200","Name":"Sales"}`); err != nil {
		t.Errorf("ReadInput should accept a JSON object, got error: %v", err)
	}
}

func TestReadInput_AcceptsArray(t *testing.T) {
	if err := readInputFromFile(t, `[{"Code":"200"},{"Code":"201"}]`); err != nil {
		t.Errorf("ReadInput should accept a JSON array, got error: %v", err)
	}
}

func TestReadInput_RejectsInvalidJSON(t *testing.T) {
	err := readInputFromFile(t, `{not valid json`)
	if err == nil {
		t.Fatal("ReadInput should reject malformed JSON")
	}
}

func TestReadInput_RejectsEmpty(t *testing.T) {
	err := readInputFromFile(t, "   ")
	if err == nil {
		t.Fatal("ReadInput should reject empty/whitespace-only input")
	}
}
