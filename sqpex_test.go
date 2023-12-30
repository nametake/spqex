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
		replace    bool
		goldenFile string
		want       *ProcessResult
	}{
		{
			filePath:   "testdata/format.go",
			command:    "xargs echo -n | sed -e 's/TABLE/TABLE_A/'",
			replace:    true,
			goldenFile: "testdata/format_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{},
				IsChanged:     true,
			},
		},
		{
			filePath:   "testdata/format.go",
			command:    "xargs echo -n | sed -e 's/TABLE/TABLE_A/'",
			replace:    false,
			goldenFile: "testdata/format_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{},
				IsChanged:     false,
			},
		},
		{
			filePath:   "testdata/has_error.go",
			command:    "./testdata/has_error.sh",
			replace:    true,
			goldenFile: "testdata/has_error_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{
					{
						Message: "COMMAND ERROR",
						PosText: "testdata/has_error.go:16:11",
					},
				},
				IsChanged: true,
			},
		},
		{
			filePath:   "testdata/has_error.go",
			command:    "./testdata/has_error.sh",
			replace:    false,
			goldenFile: "testdata/has_error_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{
					{
						Message: "COMMAND ERROR",
						PosText: "testdata/has_error.go:16:11",
					},
				},
				IsChanged: false,
			},
		},
		{
			filePath:   "testdata/error_only.go",
			command:    `echo -n "COMMAND ERROR" 1>&2 && exit 1`,
			replace:    true,
			goldenFile: "testdata/error_only_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{
					{
						Message: "COMMAND ERROR",
						PosText: "testdata/error_only.go:9:11",
					},
				},
				IsChanged: false,
			},
		},
		{
			filePath:   "testdata/error_only.go",
			command:    `echo -n "COMMAND ERROR" 1>&2 && exit 1`,
			replace:    false,
			goldenFile: "testdata/error_only_golden.go",
			want: &ProcessResult{
				ErrorMessages: []*ErrorMessage{
					{
						Message: "COMMAND ERROR",
						PosText: "testdata/error_only.go:9:11",
					},
				},
				IsChanged: false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.filePath, func(t *testing.T) {
			result, err := process(test.filePath, test.command, test.replace)
			if err != nil {
				t.Fatalf("process(%q, %q) returned unexpected error: %v", test.filePath, test.command, err)
			}

			golden, err := os.ReadFile(test.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", test.goldenFile, err)
			}
			if test.want.IsChanged {
				test.want.Output = golden
			}

			if diff := cmp.Diff(test.want, result); diff != "" {
				t.Errorf("process(%q, %q) returned unexpected result (-want +got):\n%s", test.filePath, test.command, diff)
			}
		})
	}
}

func TestFindGoFiles(t *testing.T) {
	files, err := FindGoFiles("testdata/filelist")
	if err != nil {
		t.Fatalf("findGoFiles(%q) returned unexpected error: %v", "testdata", err)
	}

	expected := []string{
		"testdata/filelist/dir/file1.go",
		"testdata/filelist/file1.go",
		"testdata/filelist/file2.go",
	}

	if diff := cmp.Diff(expected, files); diff != "" {
		t.Errorf("findGoFiles(%q) returned unexpected result (-want +got):\n%s", "testdata", diff)
	}

}
