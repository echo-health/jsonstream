# JSON Stream

Streaming parser for JSON with some simple path matching based on JSON path.

Simple example:

```go
package main

import (
	"strings"

	"github.com/echo-health/jsonstream"
	"github.com/sanity-io/litter"
)

func main() {
	p := jsonstream.New(strings.NewReader(`
        {
            "a": {
                "b": [
                    {"c": 1},
                    {"c": 2},
                    {"c": 3}
                ]
            }
        }
    `))

	// visits every element in "a.b"
	p.On("$.a.b[*]", func(key string, value struct{ C int }) error {
		litter.Dump(value)
		return nil
	})

	err := p.Decode()
	if err != nil {
		panic(err)
	}
}

```
