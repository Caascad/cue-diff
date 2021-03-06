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

	"cuelang.org/go/cue"
)

func (d *differ) diffValue(x, y cue.Value) (bool, error) {
	if d.cfg.UseDefaults {
		x, _ = x.Default()
		y, _ = y.Default()
	}

	// Return early when values have different concrete types
	if x.IncompleteKind() != y.IncompleteKind() {
		d.cl.Add(UPDATE, x.Path(), &x, &y)
		return true, nil
	}

	switch xc, yc := x.IsConcrete(), y.IsConcrete(); {

	case xc != yc:
		d.cl.Add(UPDATE, x.Path(), &x, &y)
		return true, nil

	case xc && yc:
		switch k := x.Kind(); k {
		case cue.StructKind:
			return d.diffStruct(x, y)

		case cue.ListKind:
			return d.diffList(x, y)

		}

		fallthrough

	default:
		// FIXME: to handle constraints like time.Duration that are not concrete.
		if x.Kind() == cue.BottomKind && y.Kind() == cue.BottomKind {
			if fmt.Sprint(x) != fmt.Sprint(y) {
				d.cl.Add(UPDATE, x.Path(), &x, &y)
				return true, nil
			}
			return false, nil
		}

		if !x.Equals(y) {
			d.cl.Add(UPDATE, x.Path(), &x, &y)
			return true, nil
		}
	}

	return false, nil
}
