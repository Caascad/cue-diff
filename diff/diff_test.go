// Copyright 2019 CUE Authors
// Copyright 2021 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diff

import (
	"fmt"
	"testing"

	"cuelang.org/go/cue/cuecontext"
	_ "cuelang.org/go/pkg"
	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	testCases := []struct {
		name    string
		x, y    string
		cl      testChangelog
		profile *Profile
	}{
		{
			name: "identity struct",
			x: `{
				a: {
					b: 1
					c: 2
				}
				l: {
					d: 1
				}
			}`,
			y: `{
				a: {
					c: 2
					b: 1
				}
				l: {
					d: 1
				}
			}`,
			cl: testChangelog{},
		},
		{
			name: "identity list",
			x:    `[1, 2, 3]`,
			y:    `[1, 2, 3]`,
			cl:   testChangelog{},
		},
		{
			name: "identity value",
			x:    `"foo"`,
			y:    `"foo"`,
			cl:   testChangelog{},
		},
		{
			name: "modified value",
			x:    `"foo"`,
			y:    `"bar"`,
			cl: testChangelog{
				testChange{Type: UPDATE, From: `"foo"`, To: `"bar"`},
			},
		},
		{
			name: "basics",
			x: `{
				a: int
				b: 2
				s: *4 | int
				d: 1
				e: [1, 2, 3]
				f: null
				#Def: 45
				_h: 2
			}`,
			y: `{
				a: string
				c: 3
				s: 4
				d: int
				e: {
					a: 3
				}
				f: null
				#Def: "foo"
				_h: 3
			} `,
			cl: testChangelog{
				testChange{Type: UPDATE, Path: `a`, From: `int`, To: `string`},
				testChange{Type: DELETE, Path: `b`, From: `2`, To: `<nil>`},
				testChange{Type: CREATE, Path: `c`, From: `<nil>`, To: `3`},
				testChange{Type: UPDATE, Path: `d`, From: `1`, To: `int`},
				testChange{Type: UPDATE, Path: `e`, From: `[1, 2, 3]`, To: `{
	a: 3
}`},
				testChange{Type: UPDATE, Path: `_h`, From: `2`, To: `3`},
			},
		},
		{
			name:    "ignored hidden fields",
			profile: &Profile{IgnoreHiddenFields: true, UseDefaults: true},
			x: `{
				_x: "foo"
				_y: "foo"
				xy: *_x | _y
			}`,
			y: `{
				_x: "foo"
				_y: string
				xy: *_x | _y
			}`,
			cl: testChangelog{},
		},
		{
			name: "empty",
			x:    `{a: 1}`,
			y:    `{}`,
			cl: testChangelog{
				testChange{Type: DELETE, Path: `a`, From: `1`, To: `<nil>`},
			},
		},
		{
			name: "recursion",
			x: `{
				s: {
					a: f: 3
					b: 3
					d: 4
				}
				l: [
					[3, 4]
				]
			}`,
			y: `{
				s: {
					a: f: 4
					b: 3
					c: 4
				}
				l: [
					[3, 5, 6]
				]
			}`,
			cl: testChangelog{
				testChange{Type: UPDATE, Path: `s.a.f`, From: `3`, To: `4`},
				testChange{Type: DELETE, Path: `s.d`, From: `4`, To: `<nil>`},
				testChange{Type: CREATE, Path: `s.c`, From: `<nil>`, To: `4`},
				testChange{Type: UPDATE, Path: "l[0][1]", From: "4", To: "5"},
				testChange{Type: UPDATE, Path: "l[0]", From: "[3, 4]", To: "[3, 5, 6]"},
				testChange{Type: UPDATE, Path: "l", From: "[[3, 4]]", To: "[[3, 5, 6]]"},
			},
		},
		{
			name: "bulk optional",
			x: `
				{[_]: x: "hello"}
				a: x: "hello"
				`,
			y: `[_]: x: "hello"`,
			cl: testChangelog{
				testChange{Type: DELETE, Path: `a`, From: `{
	x: "hello"
}`, To: `<nil>`},
			},
		},
		{
			name: "constraint",
			x: `
				import (
					"time"
					"list"
				)

				a: time.Duration
				b: =~ "^a"
				l: list.MaxItems(4)
				`,
			y: `
				import (
					"time"
					"list"
				)

				a: time.Duration
				b: =~"^a"
				b: "ab"
				l: list.MaxItems(5)
				`,
			cl: testChangelog{
				testChange{Type: UPDATE, Path: `b`, From: `=~"^a"`, To: `"ab"`},
				testChange{Type: UPDATE, Path: `l`, From: `list.MaxItems(4)`, To: `list.MaxItems(5)`},
			},
		},
		{
			name: "null",
			x: `
				a: [1, null, {b: null}]
				`,
			y: `
				a: [1, null, {b: null}]
				`,
			cl: testChangelog{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := cuecontext.New()
			x := ctx.CompileString(tc.x)
			if x.Err() != nil {
				t.Fatal(x.Err())
			}
			y := ctx.CompileString(tc.y)
			if y.Err() != nil {
				t.Fatal(y.Err())
			}
			p := tc.profile
			if p == nil {
				p = Final
			}
			cl, _ := p.Diff(x.Value(), y.Value())
			require.Equal(t, tc.cl, toTestChangelog(cl))
		})
	}
}

type testChangelog []testChange

type testChange struct {
	Type string
	Path string
	From string
	To   string
}

func toTestChangelog(cl Changelog) testChangelog {
	tcl := testChangelog{}
	for _, c := range cl {
		tc := testChange{
			Type: c.Type,
			Path: fmt.Sprint(c.Path),
			From: fmt.Sprint(c.From),
			To:   fmt.Sprint(c.To),
		}
		tcl = append(tcl, tc)
	}
	return tcl
}
