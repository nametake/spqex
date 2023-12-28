package main

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestProcess(t *testing.T) {
	tests := []struct {
		sourceFile string
		goldenFile string
	}{
		{
			sourceFile: "testdata/format/format.go",
			goldenFile: "testdata/format/format_golden.go",
		},
	}

	for _, test := range tests {
		t.Run(test.sourceFile, func(t *testing.T) {
			result, err := process(test.sourceFile)
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
