package keyvadb

import (
	"fmt"
	"math"
	"strings"
)

type Level struct {
	Nodes      uint64
	Entries    uint64
	Synthetics uint64
	WellFormed uint64
}

func (l Level) String() string {
	wellFormed := float64(l.WellFormed) / float64(l.Nodes) * 100
	synthetics := float64(l.Synthetics) / float64(l.Entries) * 100
	return fmt.Sprintf("Nodes (good): %8d (%6.2f%%)\tEntries (synthetic): %8d (%6.2f%%)", l.Nodes, wellFormed, l.Entries, synthetics)
}

func (l *Level) Add(b Level) {
	l.Nodes += b.Nodes
	l.Entries += b.Entries
	l.Synthetics += b.Synthetics
	l.WellFormed += b.WellFormed
}

func (l *Level) Merge(n *Node, depth uint64) {
	l.Nodes++
	l.Entries += uint64(n.Occupancy())
	l.Synthetics += uint64(n.Synthetics())
	if n.SanityCheck() {
		l.WellFormed++
	}
}

func (l Level) NonSyntheticEntries() uint64 {
	return l.Entries - l.Synthetics
}

type Summary struct {
	Total  Level
	Levels []Level
	Degree uint64
}

func NewSummary(tree *Tree) (*Summary, error) {
	sum := &Summary{
		Degree: tree.Degree,
	}
	err := tree.Each(func(level int, n *Node) error {
		if level >= len(sum.Levels) {
			sum.Levels = append(sum.Levels, Level{})
		}
		sum.Levels[level].Merge(n, uint64(level))
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, level := range sum.Levels {
		sum.Total.Add(level)
	}
	return sum, nil
}

func (sum Summary) MaxNodes(depth int) uint64 {
	return uint64(math.Pow(float64(sum.Degree), float64(depth)))
}

func (sum Summary) Overall() string {
	averageEntries := float64(sum.Total.NonSyntheticEntries()) / float64(sum.Total.Nodes)
	efficency := averageEntries / float64(sum.Degree) * 100
	return fmt.Sprintf("Total:\t\t%s\tReal Entries/Node: %6.2f\t Efficiency: %6.2f%%", sum.Total, averageEntries, efficency)
}

func (sum Summary) String() string {
	var s []string
	for i, level := range sum.Levels {
		occupied := float64(level.Nodes) / float64(sum.MaxNodes(i)) * 100
		share := float64(level.Nodes) / float64(sum.Total.Nodes) * 100
		s = append(s, fmt.Sprintf("Level: %3d\t%s\tOccupied: %6.2f%%\tShare: %6.2f%%", i, level, occupied, share))
	}
	s = append(s, sum.Overall())
	return strings.Join(s, "\n")
}
