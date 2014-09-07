package keyva

import "fmt"

type Target struct {
	Index int
	Key   Hash
}

type TargetSlice []Target

func NewTargetSlice(start, end Hash, keys HashSlice) TargetSlice {
	stride := start.Stride(end, int64(len(keys)))
	cursor := start.Add(stride)
	var targets TargetSlice
	for i, key := range keys {
		if key == EmptyItem {
			targets = append(targets, Target{Index: i, Key: cursor})
		}
		cursor = cursor.Add(stride)
	}
	return targets
}

func (t Target) String() string {
	return fmt.Sprintf("%s:%d", t.Key, t.Index)
}

func (t TargetSlice) String() string {
	return dumpWithTitle("Targets", t, 0)
}
