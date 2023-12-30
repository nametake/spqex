package spqex

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestProcess(t *testing.T) {
	tests := []struct {
		filePath   string
		command    string
		goldenFile string
		want       *ProcessResult
	}{
		{
			filePath:   "testdata/format/format.go",
			command:    "xargs echo -n | sed -e 's/TABLE/TABLE_A/'",
			goldenFile: "testdata/format/format_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{},
				IsChanged:     true,
			},
		},
		{
			filePath:   "testdata/format/has_error.go",
			command:    "./testdata/format/has_error.sh",
			goldenFile: "testdata/format/has_error_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{
					{Message: "COMMAND ERROR"},
				},
				IsChanged: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.filePath, func(t *testing.T) {
			result, err := process(test.filePath, test.command)
			if err != nil {
				t.Fatalf("process(%q, %q) returned unexpected error: %v", test.filePath, test.command, err)
			}

			golden, err := os.ReadFile(test.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", test.goldenFile, err)
			}
			test.want.Output = golden

			if diff := cmp.Diff(test.want, result); diff != "" {
				t.Errorf("process(%q, %q) returned unexpected result (-want +got):\n%s", test.filePath, test.command, diff)
			}
		})
	}
}
