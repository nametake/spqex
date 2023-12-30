package format

import (
	"cloud.google.com/go/spanner"
)

func NoError() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE_A;",
		Params: map[string]interface{}{},
	}
}

func HasError() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM HAS_ERROR;",
		Params: map[string]interface{}{},
	}
}
