package keyvadb

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/dustin/go-humanize"
)

type Cache struct {
	m         map[NodeId]*list.Element
	l         *list.List
	size      int
	hits      uint64
	misses    uint64
	evictions uint64
	sync.RWMutex
}

func NewCache(size int) *Cache {
	return &Cache{
		m:    make(map[NodeId]*list.Element, size),
		l:    list.New(),
		size: size,
	}
}

func (c *Cache) Get(id NodeId) *Node {
	c.RLock()
	defer c.RUnlock()
	if e, ok := c.m[id]; ok {
		c.hits++
		c.l.MoveToFront(e)
		return e.Value.(*Node)
	}
	c.misses++
	return nil
}

func (c *Cache) Set(node *Node) {
	c.Lock()
	defer c.Unlock()
	if e, ok := c.m[node.Id]; ok {
		e.Value = node
		c.l.MoveToFront(e)
		return
	}
	c.m[node.Id] = c.l.PushFront(node)
	if c.l.Len() > c.size {
		c.evictions++
		last := c.l.Back()
		c.l.Remove(last)
		delete(c.m, last.Value.(*Node).Id)
	}
}

func (c *Cache) String() string {
	c.RLock()
	defer c.RUnlock()
	hits := float64(c.hits) / float64(c.hits+c.misses) * 100
	return fmt.Sprintf("Cache Size: %s Hits: %0.2f%% Evictions: %d", humanize.Comma(int64(len(c.m))), hits, c.evictions)
}
