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
	"cuelang.org/go/cue"
)

func (d *differ) diffStruct(x, y cue.Value) (bool, error) {
	hasDiff := false
	sx, _ := x.Struct()
	sy, _ := y.Struct()

	// Best-effort topological sort, prioritizing x over y, using a variant of
	// Kahn's algorithm (see, for instance
	// https://www.geeksforgeeks.org/topological-sorting-indegree-based-solution/).
	// We assume that the order of the elements of each value indicate an edge
	// in the graph. This means that only the next unprocessed nodes can be
	// those with no incoming edges.
	xMap := make(map[string]int32, sx.Len())
	yMap := make(map[string]int32, sy.Len())
	for i := 0; i < sx.Len(); i++ {
		xMap[sx.Field(i).Selector] = int32(i + 1)
	}
	for i := 0; i < sy.Len(); i++ {
		yMap[sy.Field(i).Selector] = int32(i + 1)
	}

	var xi, yi int
	var xf, yf cue.FieldInfo

	for xi < sx.Len() || yi < sy.Len() {
		// Process zero nodes
		for ; xi < sx.Len(); xi++ {
			xf = sx.Field(xi)
			xv := xf.Value
			yp := yMap[xf.Selector]
			if yp > 0 {
				break
			}
			if xf.IsOptional && d.cfg.IgnoreOptionalFields {
				continue
			}
			if xf.IsDefinition && d.cfg.IgnoreDefinitions {
				continue
			}
			if xf.IsHidden && d.cfg.IgnoreHiddenFields {
				continue
			}
			// Add subfields to the changelog as well
			xv.Walk(func(v cue.Value) bool {
				d.cl.Add(DELETE, v.Path(), &v, nil)
				return true
			}, nil)
			hasDiff = true
		}
		for ; yi < sy.Len(); yi++ {
			yf = sy.Field(yi)
			yv := yf.Value
			if yMap[yf.Selector] == 0 {
				// already done
				continue
			}
			xp := xMap[yf.Selector]
			if xp > 0 {
				break
			}
			yMap[yf.Selector] = 0

			if yf.IsOptional && d.cfg.IgnoreOptionalFields {
				continue
			}
			if yf.IsDefinition && d.cfg.IgnoreDefinitions {
				continue
			}
			if yf.IsHidden && d.cfg.IgnoreHiddenFields {
				continue
			}
			// Add subfields to the changelog as well
			yv.Walk(func(v cue.Value) bool {
				d.cl.Add(CREATE, v.Path(), nil, &v)
				return true
			}, nil)
			hasDiff = true
		}

		// Compare nodes
		for ; xi < sx.Len(); xi++ {
			xf = sx.Field(xi)

			yp := yMap[xf.Selector]
			if yp == 0 {
				break
			}
			// If yp != xi+1, the topological sort was not possible.
			yMap[xf.Selector] = 0

			yf := sy.Field(int(yp - 1))

			xv := xf.Value
			yv := yf.Value

			if d.cfg.IgnoreOptionalFields {
				if xf.IsOptional && yf.IsOptional {
					continue
				} else if xf.IsOptional {
					// Behave like x doesn't exists.
					yv.Walk(func(v cue.Value) bool {
						d.cl.Add(CREATE, v.Path(), nil, &v)
						return true
					}, nil)
					continue
				}
			}

			if xf.IsDefinition && d.cfg.IgnoreDefinitions {
				continue
			}
			if xf.IsHidden && d.cfg.IgnoreHiddenFields {
				continue
			}

			hasDiff, err := d.diffValue(xv, yv)
			if err != nil {
				return hasDiff, err
			}

		}
	}

	return hasDiff, nil
}
