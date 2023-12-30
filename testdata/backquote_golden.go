package format

import (
	"cloud.google.com/go/spanner"
)

func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE_A;",
		Params: map[string]interface{}{},
	}
}
