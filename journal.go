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

func (m *SimpleJournal) Swap(current, previous *Node) {
	m.current = append(m.current, current)
	m.previous = append(m.previous, previous)
}

func (m *SimpleJournal) Commit() error {
	for _, current := range m.current {
		if err := m.keys.Set(current); err != nil {
			return err
		}
	}
	m.current = m.current[:0]
	m.previous = m.previous[:0]
	return nil
}

func (m *SimpleJournal) String() string {
	var s []string
	for i := range m.current {
		s = append(s, fmt.Sprintf("%08d:%03d:%03d", m.current[i].Id, m.previous[i].Occupancy(), m.current[i].Occupancy()))
	}
	return dumpWithTitle("Journal", s, 0)
}

func NewFileJournal(name string, keys KeyStore, values ValueStore) (*FileJournal, error) {
	journal := NewSimpleJournal(name, keys, values)
	return &FileJournal{journal}, nil
}

type FileJournal struct {
	*SimpleJournal
}

func (f *FileJournal) Commit() error {
	return f.SimpleJournal.Commit()
}
