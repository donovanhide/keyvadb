package keyvadb

import "fmt"

type KeyValue struct {
	Key
	Value []byte
}
type KeyValueSlice []KeyValue

func (s KeyValueSlice) Keys() KeySlice {
	var k KeySlice
	for _, kv := range s {
		k = append(k, kv.Key)
	}
	return k
}

func (kv KeyValue) String() string {
	return fmt.Sprintf("%s:%X", kv.Key, kv.Value)
}
