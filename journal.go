package keyvadb

import "fmt"

type Delta struct {
	current  *Node
	previous *Node
}

func (d *Delta) NewKeys() int {
	return d.current.Occupancy() - d.previous.Occupancy()
}

func (d *Delta) NewChildren() int {
	return d.current.ChildCount() - d.previous.ChildCount()
}

func (d *Delta) String() string {
	return fmt.Sprintf("%016d New Keys: %03d: New Children: %03d", d.current.Id, d.NewKeys(), d.NewChildren())
}

func NewSimpleJournal(name string, keys KeyStore, values ValueStore) *SimpleJournal {
	return &SimpleJournal{
		name:   name,
		keys:   keys,
		values: values,
	}
}

type SimpleJournal struct {
	name   string
	keys   KeyStore
	values ValueStore
	deltas []Delta
}

func (j *SimpleJournal) Len() int {
	return len(j.deltas)
}

func (j *SimpleJournal) Swap(current, previous *Node) {
	j.deltas = append(j.deltas, Delta{current, previous})
}

func (j *SimpleJournal) Commit() error {
	for _, delta := range j.deltas {
		if err := j.keys.Set(delta.current); err != nil {
			return err
		}
	}
	j.deltas = nil
	return nil
}

func (j *SimpleJournal) String() string {
	return dumpWithTitle("Journal", j.deltas, 0)
}

func (j *SimpleJournal) Close() error {
	return nil
}

func NewFileJournal(name string, keys KeyStore, values ValueStore) (*FileJournal, error) {
	journal := NewSimpleJournal(name, keys, values)
	return &FileJournal{journal}, nil
}

type FileJournal struct {
	*SimpleJournal
}

func (j *FileJournal) Close() error {
	return nil
}

func (f *FileJournal) Commit() error {
	if err := f.SimpleJournal.Commit(); err != nil {
		return err
	}
	return f.keys.Sync()
}
