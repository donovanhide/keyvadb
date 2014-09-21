package keyvadb

import "sync"

type Buffer struct {
	m map[Hash]*Key
	sync.RWMutex
}

func NewBuffer(size uint64) *Buffer {
	return &Buffer{
		m: make(map[Hash]*Key, int(size)),
	}
}

func (b *Buffer) Add(key *Key) uint64 {
	b.RLock()
	b.m[key.Hash] = key
	length := uint64(len(b.m))
	b.RUnlock()
	return length
}

func (b *Buffer) Get(hash Hash) *Key {
	b.RLock()
	key := b.m[hash]
	b.RUnlock()
	return key
}

func (b *Buffer) Keys() KeySlice {
	var keys KeySlice
	b.RLock()
	for _, key := range b.m {
		keys = append(keys, *key)
	}
	b.RUnlock()
	return keys
}

func (b *Buffer) Remove(keys KeySlice) {
	b.Lock()
	for _, key := range keys {
		delete(b.m, key.Hash)
	}
	b.Unlock()
}
