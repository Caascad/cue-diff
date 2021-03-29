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

	"cuelang.org/go/cue"
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
				#Def: 45
			}`,
			y: `{
				a: string
				c: 3
				s: 4
				d: int
				e: {
					a: 3
				}
				#Def: "foo"
			} `,
			cl: testChangelog{
				testChange{Type: UPDATE, Path: `a`, From: `int`, To: `string`},
				testChange{Type: DELETE, Path: `b`, From: `2`, To: `<nil>`},
				testChange{Type: CREATE, Path: `c`, From: `<nil>`, To: `3`},
				testChange{Type: UPDATE, Path: `d`, From: `1`, To: `int`},
				testChange{Type: UPDATE, Path: `e`, From: `[1, 2, 3]`, To: `{
	a: 3
}`},
			},
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
				testChange{Type: UPDATE, Path: `l`, From: `[[3, 4]]`, To: `[[3, 5, 6]]`},
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var r cue.Runtime
			x, err := r.Compile("x", tc.x)
			if err != nil {
				t.Fatal(err)
			}
			y, err := r.Compile("y", tc.y)
			if err != nil {
				t.Fatal(err)
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
