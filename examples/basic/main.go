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
