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
