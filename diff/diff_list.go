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

func (d *differ) diffList(x, y cue.Value) error {
	ix, _ := x.List()
	iy, _ := y.List()

	for {
		hasX := ix.Next()
		hasY := iy.Next()

		if !hasX && !hasY {
			return nil
		}

		if !hasX && hasY || hasX && !hasY {
			d.cl.Add(UPDATE, x.Path(), &x, &y)
			return nil
		}

		if !ix.Value().Equals(iy.Value()) {
			d.cl.Add(UPDATE, x.Path(), &x, &y)
			return nil
		}
	}
}
