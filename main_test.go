package main

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestProcess(t *testing.T) {
	tests := []struct {
		sourceFile string
		command    string
		goldenFile string
	}{
		{
			sourceFile: "testdata/format/format.go",
			command:    "xargs echo -n | gsed -e 's/TABLE/TABLE_A/'",
			goldenFile: "testdata/format/format_golden.go",
		},
	}

	for _, test := range tests {
		t.Run(test.sourceFile, func(t *testing.T) {
			result, err := process(test.sourceFile, test.command)
			if err != nil {
				t.Fatalf("process(%q) returned error %v", test.sourceFile, err)
			}

			golden, err := os.ReadFile(test.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", test.goldenFile, err)
			}

			if diff := cmp.Diff(string(golden), string(result)); diff != "" {
				t.Errorf("process(%q) returned unexpected result (-want +got):\n%s", test.sourceFile, diff)
			}
		})
	}
}
