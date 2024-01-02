# spqex

spqex is a tool that extracts SQL from the `spanner.Statement` structure in the `cloud.google.com/go/spanner` package in Go and executes the specified command.

It takes the extracted SQL and executes the specified command with it as standard input.

spqex has two modes: fmt and lint.
In fmt mode, if the command succeeds, it replaces the SQL in the standard output result.
In lint mode, no replacement is performed.
In either mode, if the executed command fails, spqex displays the content of standard error, and it is considered a failure.

## Installation

```console
go install github.com/nametake/spqex/cmd/spqex@latest
```

## Usage

```
Usage: spqex [options] directory
Options:
  -cmd string
        Specify command to execute
  -mode string
        Specify mode (lint or fmt). default: lint (default "lint")
```

## Example

The following is an example using [sql-formatter](https://github.com/sql-formatter-org/sql-formatter).

The original Go code before execution is as follows:

```go
package main

import (
	"cloud.google.com/go/spanner"
)

func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL:    "SELECT * FROM TABLE;",
		Params: map[string]interface{}{},
	}
}
```

By running the following command in the directory where the code is located, the SQL in the Go code is replaced with formatted SQL using sql-formatter.

```console
spqex -mode fmt -cmd 'sql-formatter --language bigquery' .
```

```go
package main

import (
	"cloud.google.com/go/spanner"
)

func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL: `
SELECT
  *
FROM
  TABLE;
`, Params: map[string]interface{}{},
	}
}
```

## Note

If you want to dynamically use ORDER BY with cloud.google.com/go/spanner, a [method using fmt.Sprintf](https://github.com/googleapis/google-cloud-go/issues/6496) has been proposed.

Therefore, to handle cases like the one below in spqex, it replaces format verbs before passing them to the command.

```go
func SQL() *spanner.Statement {
	return &spanner.Statement{
		SQL:    fmt.Sprintf("SELECT * FROM TABLE ORDER BY %s;", "CreatedAt"),
		Params: map[string]interface{}{},
	}
}
```

The SQL passed to the command is converted as follows:

```
SELECT * FROM TABLE ORDER BY _DUMMY_STRING_;
```

The conversion table is as follows:

| Verbs | Dummy String     |
| ---   | ---              |
| `%s`  | `_DUMMY_STRING_` |
| `%v`  | `_DUMMY_VALUE_`  |
| `%d`  | `-999`           |

In fmt mode, the replaced strings are processed to revert them to format verbs, so if the original SQL contains strings from the conversion table, unintended behavior may occur.
