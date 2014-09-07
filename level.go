package keyva

import (
	"fmt"
	"strings"
)

type Level struct {
	Nodes      int
	Entries    int
	WellFormed int
}

func (l Level) String() string {
	return fmt.Sprintf("Nodes: %-8d Entries: %-8d WellFormed: %-8d", l.Nodes, l.Entries, l.WellFormed)
}

func (l *Level) Add(b Level) {
	l.Nodes += b.Nodes
	l.Entries += b.Entries
	l.WellFormed += b.WellFormed
}

func (l *Level) Merge(n *Node) {
	l.Nodes++
	l.Entries += n.Occupancy()
	if n.SanityCheck() {
		l.WellFormed++
	}
}

type LevelSlice []Level

func (l LevelSlice) Total() Level {
	var total Level
	for _, level := range l {
		total.Add(level)
	}
	return total
}

func (l LevelSlice) String() string {
	total := l.Total()
	var s []string
	for i, level := range l {
		s = append(s, fmt.Sprintf("Level: %-3d %s Share %0.2f%%", i, level, float64(level.Nodes)/float64(total.Nodes)*100))
	}
	s = append(s, "Total:     "+total.String())
	return strings.Join(s, "\n")
}

func (s *LevelSlice) Add(n *Node, level int) {
	if len(*s) <= level {
		*s = append(*s, Level{})
	}
	(*s)[level].Merge(n)
}
