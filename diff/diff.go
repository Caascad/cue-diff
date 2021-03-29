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

const (
	// CREATE represents when an element has been added
	CREATE = "create"
	// UPDATE represents when an element has been updated
	UPDATE = "update"
	// DELETE represents when an element has been removed
	DELETE = "delete"
)

// Changelog stores a list of changed items
type Changelog []Change

func (cl *Changelog) Add(t string, path cue.Path, from *cue.Value, to *cue.Value) {
	change := Change{
		Type: t,
		Path: path,
		From: from,
		To:   to,
	}
	(*cl) = append((*cl), change)
}

// Change stores information about a changed item
type Change struct {
	Type string
	Path cue.Path
	From *cue.Value
	To   *cue.Value
}

type Profile struct {
	UseDefaults       bool
	IgnoreDefinitions bool
}

var (
	// Schema is the standard profile used for comparing schema.
	Schema = &Profile{}

	// Final is the standard profile for comparing data.
	Final = &Profile{
		UseDefaults:       true,
		IgnoreDefinitions: true,
	}
)

// Diff is a shorthand for Final.Diff.
func Diff(x, y cue.Value) (Changelog, error) {
	return Final.Diff(x, y)
}

// Diff returns an edit script representing the difference between x and y.
func (p *Profile) Diff(x, y cue.Value) (Changelog, error) {
	d := differ{
		cfg: *p,
		cl:  make([]Change, 0),
	}
	return d.cl, d.diffValue(x, y)
}

type differ struct {
	cfg Profile
	cl  Changelog
}
