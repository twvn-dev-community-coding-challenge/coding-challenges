// Package passenger holds the seeded passenger records. Registration is out
// of scope, so the directory is read-only after start-up.
package passenger

import "context"

type Directory struct {
	ids map[string]bool
}

func NewDirectory(ids ...string) *Directory {
	d := &Directory{ids: map[string]bool{}}
	for _, id := range ids {
		d.ids[id] = true
	}
	return d
}

func (d *Directory) Exists(_ context.Context, id string) bool { return d.ids[id] }
