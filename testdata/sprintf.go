package format

import (
	"fmt"

	"cloud.google.com/go/spanner"
)

func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL:    fmt.Sprintf("SELECT * FROM TABLE ORDER BY %s;", "CreatedAt"),
		Params: map[string]interface{}{},
	}
}
