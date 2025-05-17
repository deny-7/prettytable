# prettytable (Go)

A Go port of the Python PrettyTable package. Display tabular data in a visually appealing ASCII table format.

## Features
- Create tables with custom field (column) names
- Add rows one by one or all at once
- Print table as ASCII

## Installation

```
go get github.com/deny-7/prettytable
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/deny-7/prettytable"
)

func main() {
	t := prettytable.NewTableWithFields([]string{"City name", "Area", "Population", "Annual Rainfall"})
	t.AddRows([][]any{
		{"Adelaide", 1295, 1158259, 600.5},
		{"Brisbane", 5905, 1857594, 1146.4},
		{"Darwin", 112, 120900, 1714.7},
		{"Hobart", 1357, 205556, 619.5},
		{"Sydney", 2058, 4336374, 1214.8},
		{"Melbourne", 1566, 3806092, 646.9},
		{"Perth", 5386, 1554769, 869.4},
	})
	fmt.Println(t)
}
```

## Example Output

```
+-----------+------+------------+-----------------+
| City name | Area | Population | Annual Rainfall |
+-----------+------+------------+-----------------+
| Adelaide  | 1295 | 1158259    | 600.5           |
| Brisbane  | 5905 | 1857594    | 1146.4          |
| Darwin    | 112  | 120900     | 1714.7          |
| Hobart    | 1357 | 205556     | 619.5           |
| Sydney    | 2058 | 4336374    | 1214.8          |
| Melbourne | 1566 | 3806092    | 646.9           |
| Perth     | 5386 | 1554769    | 869.4           |
+-----------+------+------------+-----------------+
```

## License
MIT
