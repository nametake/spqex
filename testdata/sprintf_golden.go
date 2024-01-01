package format

import (
	"fmt"

	"cloud.google.com/go/spanner"
)

func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL:    fmt.Sprintf("SELECT * FROM TABLE_A ORDER BY _DUMMY_STRING_;", "CreatedAt"),
		Params: map[string]interface{}{},
	}
}
