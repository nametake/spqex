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
				File:          "testdata/format.go",
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
				File:          "testdata/format.go",
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
				File: "testdata/has_error.go",
				ErrorMessages: []*ErrorMessage{
					{
						Query:   "SELECT * FROM HAS_ERROR;",
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
				File: "testdata/has_error.go",
				ErrorMessages: []*ErrorMessage{
					{
						Query:   "SELECT * FROM HAS_ERROR;",
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
				File: "testdata/error_only.go",
				ErrorMessages: []*ErrorMessage{
					{
						Query:   "SELECT * FROM TABLE;",
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
				File: "testdata/error_only.go",
				ErrorMessages: []*ErrorMessage{
					{
						Query:   "SELECT * FROM TABLE;",
						Message: "COMMAND ERROR",
						PosText: "testdata/error_only.go:9:11",
					},
				},
				IsChanged: false,
			},
		},
		{
			filePath:   "testdata/multiline.go",
			command:    "xargs echo -n | sed -e 's/TABLE;/\\nTABLE_A;\\n/'",
			replace:    true,
			goldenFile: "testdata/multiline_golden.go",
			want: &ProcessResult{
				File:          "testdata/multiline.go",
				ErrorMessages: []*ErrorMessage{},
				IsChanged:     true,
			},
		},
		{
			filePath:   "testdata/backquote.go",
			command:    "xargs echo -n | sed -e 's/TABLE/TABLE_A/'",
			replace:    true,
			goldenFile: "testdata/backquote_golden.go",
			want: &ProcessResult{
				File:          "testdata/backquote.go",
				ErrorMessages: []*ErrorMessage{},
				IsChanged:     true,
			},
		},
		{
			filePath:   "testdata/sprintf.go",
			command:    "xargs echo -n | sed -e 's/TABLE/TABLE_A/'",
			replace:    true,
			goldenFile: "testdata/sprintf_golden.go",
			want: &ProcessResult{
				File:          "testdata/sprintf.go",
				ErrorMessages: []*ErrorMessage{},
				IsChanged:     true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.filePath, func(t *testing.T) {
			result, err := Process(test.filePath, test.command, test.replace)
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

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  `"SELECT * FROM TABLE_A;"`,
			want: `SELECT * FROM TABLE_A;`,
		},
		{
			arg:  "`SELECT * FROM TABLE_A;`",
			want: `SELECT * FROM TABLE_A;`,
		},
	}

	for _, test := range tests {
		t.Run(test.arg, func(t *testing.T) {
			got := trimQuotes(test.arg)
			if got != test.want {
				t.Errorf("trimQuotes(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}

func TestFillFormatVerbs(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  "SELECT * FROM TABLE ORDER BY %s;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %v;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %d;",
			want: "SELECT * FROM TABLE ORDER BY -999;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %v %v;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_VALUE_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %s %s;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_ _DUMMY_STRING_;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY %v %s;",
			want: "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_STRING_;",
		},
	}
	for _, test := range tests {
		t.Run(test.arg, func(t *testing.T) {
			got := fillFormatVerbs(test.arg)
			if got != test.want {
				t.Errorf("fillFormatVerbs(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}

func TestRestoreFormatVerbs(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_;",
			want: "SELECT * FROM TABLE ORDER BY %s;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_;",
			want: "SELECT * FROM TABLE ORDER BY %v;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY -999;",
			want: "SELECT * FROM TABLE ORDER BY %d;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_VALUE_;",
			want: "SELECT * FROM TABLE ORDER BY %v %v;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_STRING_ _DUMMY_STRING_;",
			want: "SELECT * FROM TABLE ORDER BY %s %s;",
		},
		{
			arg:  "SELECT * FROM TABLE ORDER BY _DUMMY_VALUE_ _DUMMY_STRING_;",
			want: "SELECT * FROM TABLE ORDER BY %v %s;",
		},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			got := restoreFormatVerbs(test.arg)
			if got != test.want {
				t.Errorf("restoreFormatVerbs(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}

func TestRemoveNewlines(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{
			arg:  "SELECT * FROM TABLE;",
			want: "SELECT * FROM TABLE;",
		},
		{
			arg: `
SELECT
  *
FROM
  TABLE;
`,
			want: "SELECT * FROM TABLE;",
		},
		{
			arg: `
SELECT
  *
FROM
  TABLE
ORDER BY
  CreatedAt;
`,
			want: "SELECT * FROM TABLE ORDER BY CreatedAt;",
		},
	}

	for _, test := range tests {
		t.Run(test.arg, func(t *testing.T) {
			got := removeNewlines(test.arg)
			if got != test.want {
				t.Errorf("removeNewlines(%q) = %q, want %q", test.arg, got, test.want)
			}
		})
	}
}
