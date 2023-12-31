package format

import (
	"cloud.google.com/go/spanner"
)

func SQL1() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE;",
		Params: map[string]interface{}{},
	}
}

func SQL2() *spanner.Statement {
	return &spanner.Statement{
		SQL: `
SELECT * FROM
TABLE;
`, Params: map[string]interface{}{},
	}
}

func SQL3() *spanner.Statement {
	return &spanner.Statement{
		SQL: `
SELECT * FROM
TABLE;
`,
		Params: map[string]interface{}{},
	}
}
