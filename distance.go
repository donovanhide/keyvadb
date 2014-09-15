package keyvadb

type Distance struct {
	Distance Hash
	Key      Key
}

type DistanceMap map[int]Distance

func (s KeySlice) FindNearestKeys(n *Node) DistanceMap {
	start := n.Start.Big()
	stride := n.Stride().Big()
	halfStride := n.Stride().Divide(2).Big()
	entries := int64(n.MaxEntries())
	nearest := make(DistanceMap)
	for _, key := range s {
		index, distance := key.Hash.NearestStride(start, stride, halfStride, entries)
		if d, ok := nearest[index]; !ok || distance.Less(d.Distance) {
			nearest[index] = Distance{distance, key}
		}
	}
	return nearest
}
