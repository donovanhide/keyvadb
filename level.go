package keyva

import (
	"fmt"
	"math"
	"strings"
)

type Level struct {
	Nodes      int
	Entries    int
	WellFormed int
}

func (l Level) String() string {
	return fmt.Sprintf("Nodes: %8d\tEntries: %8d\tWellFormed: %8d", l.Nodes, l.Entries, l.WellFormed)
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
	var sumExpected float64
	var s []string
	for i, level := range l {
		expected := math.Pow(float64(ChildCount), float64(i))
		sumExpected += expected
		occupied := float64(level.Nodes) / expected * 100
		share := float64(level.Nodes) / float64(total.Nodes) * 100
		s = append(s, fmt.Sprintf("Level: %3d\t%s\tOccupied: %6.2f%%\tShare: %6.2f%%", i, level, occupied, share))
	}
	averageEntries := float64(total.Entries) / float64(total.Nodes)
	efficency := averageEntries / float64(ItemCount) * 100
	s = append(s, fmt.Sprintf("Total:\t\t%s\tEntries/Node: %6.2f\t Efficiency: %6.2f%%", total, averageEntries, efficency))
	return strings.Join(s, "\n")
}

func (s *LevelSlice) Add(n *Node, level int) {
	if len(*s) <= level {
		*s = append(*s, Level{})
	}
	(*s)[level].Merge(n)
}
