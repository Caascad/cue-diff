# cue-diff

A library for diffing cue values.

Similar to https://github.com/r3labs/diff `cue-diff` returns a changelog of
detected changes:

```go
type Changelog []Change

type Change struct {
	Type string
	Path cue.Path
	From *cue.Value
	To   *cue.Value
}
```

## Usage

### Basic example

```go
package main

import (
	"fmt"

	"cuelang.org/go/cue/cuecontext"
	"github.com/Caascad/cue-diff/diff"
)

func main() {
	ctx := cuecontext.New()

	v1 := ctx.CompileString(`{
    a: 1
    b: "foo"
}`)
	v2 := ctx.CompileString(`{
a: 2
}`)

	cl, _ := diff.Diff(v1, v2)
	fmt.Println(fmt.Sprintf("%+v", cl))
}
```

```sh
 go run main.go
[{Type:update Path:a From:1 To:2} {Type:delete Path:b From:"foo" To:<nil>}]
```

### Options

It is possible to skip the diffing of specific cue fields (definitions, hidden
fields, optional fields) by configuring a `Profile`.

```go
package main

import (
	"fmt"

	"cuelang.org/go/cue/cuecontext"
	"github.com/Caascad/cue-diff/diff"
)

func main() {
	ctx := cuecontext.New()

	v1 := ctx.CompileString(`{
    #A: {
        a: >0
    }
    _v: 1
    s: #A & {
        a: _v
    }
}`)
	v2 := ctx.CompileString(`{
    #A: {
        a: >2
    }
    _v: 3
    s: #A & {
        a: _v
    }
}`)

	p := &diff.Profile{
		IgnoreHiddenFields: true,
		IgnoreDefinitions:  true,
	}
	cl, _ := p.Diff(v1, v2)
	fmt.Println(fmt.Sprintf("%+v", cl))
}
```

```sh
 go run main.go
[{Type:update Path:s.a From:1 To:3}]
```
