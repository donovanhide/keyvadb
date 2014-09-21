package keyvadb

import "fmt"

func NewSimpleJournal(name string, keys KeyStore, values ValueStore) *SimpleJournal {
	return &SimpleJournal{
		name:   name,
		keys:   keys,
		values: values,
	}
}

type SimpleJournal struct {
	name     string
	keys     KeyStore
	values   ValueStore
	current  []*Node
	previous []*Node
}

func (j *SimpleJournal) Len() int {
	return len(j.current)
}

func (j *SimpleJournal) Swap(current, previous *Node) {
	j.current = append(j.current, current)
	j.previous = append(j.previous, previous)
}

func (j *SimpleJournal) Commit() error {
	for _, current := range j.current {
		if err := j.keys.Set(current); err != nil {
			return err
		}
	}
	j.current = j.current[:0]
	j.previous = j.previous[:0]
	return nil
}

func (j *SimpleJournal) String() string {
	var s []string
	for i := range j.current {
		keyDelta := j.current[i].Occupancy() - j.previous[i].Occupancy()
		childDelta := j.current[i].ChildCount() - j.previous[i].ChildCount()
		s = append(s, fmt.Sprintf("%016d New Keys: %03d: New Children: %03d", j.current[i].Id, keyDelta, childDelta))
	}
	return dumpWithTitle("Journal", s, 0)
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
