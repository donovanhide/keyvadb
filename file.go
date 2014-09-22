package keyvadb

import (
	"fmt"
	"io"
	"math"
	"os"
	"sync/atomic"

	"github.com/dustin/go-humanize"
	"github.com/siddontang/go/ioutil2"
)

func NewFileKeyStore(degree, cacheLevels uint64, filename string) (KeyStore, error) {
	f, err := os.OpenFile(filename+".keys", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size()%NodeBlockSize != 0 {
		// TODO: truncate instead?
		return nil, fmt.Errorf("Corrupt key store")
	}
	// Sum of consecutive powers of degree
	cacheSize := int(math.Pow(float64(degree), float64(cacheLevels)) - 1/(float64(degree)-1))
	return &FileKeyStore{
		f:      f,
		length: fi.Size(),
		cache:  NewCache(cacheSize),
	}, nil
}

type FileKeyStore struct {
	f      *os.File
	length int64
	cache  *Cache
}

func (s *FileKeyStore) Length() int64 {
	return atomic.LoadInt64(&s.length)
}

func (s *FileKeyStore) String() string {
	return fmt.Sprintf("Keys: %s %s", humanize.Bytes(uint64(s.Length())), s.cache)
}

func (s *FileKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	offset := atomic.AddInt64(&s.length, NodeBlockSize)
	debugPrintln("File Key New:", offset)
	node := NewNode(start, end, NodeId(offset), degree)
	return node, nil
}

func (s *FileKeyStore) Get(id NodeId, degree uint64) (*Node, error) {
	if node := s.cache.Get(id); node != nil {
		return node, nil
	}
	node := NewNode(FirstHash, LastHash, id, degree)
	debugPrintln("File Key Get:", id)
	r := io.NewSectionReader(s.f, int64(id), NodeBlockSize)
	if _, err := node.ReadFrom(r); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *FileKeyStore) Set(node *Node) error {
	debugPrintln("File Key Set:", node.Id)
	s.cache.Set(node)
	w := ioutil2.NewSectionWriter(s.f, int64(node.Id), NodeBlockSize)
	_, err := node.WriteTo(w)
	return err
}

func (s *FileKeyStore) Close() error {
	if err := s.f.Sync(); err != nil {
		return err
	}
	return s.f.Close()
}

func (s *FileKeyStore) Sync() error {
	return s.f.Sync()
}

func NewFileValueStore(filename string) (ValueStore, error) {
	f, err := os.OpenFile(filename+".values", os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &FileValueStore{
		f:      f,
		length: fi.Size(),
	}, nil
}

type FileValueStore struct {
	f      *os.File
	length int64
}

func (s *FileValueStore) Length() int64 {
	return atomic.LoadInt64(&s.length)
}

func (s *FileValueStore) String() string {
	return fmt.Sprintf("Values: %s", humanize.Bytes(uint64(s.Length())))
}

func (s *FileValueStore) Append(key Hash, value []byte) (*KeyValue, error) {
	size := int64(SizeOfKeyValue(value))
	length := atomic.AddInt64(&s.length, size)
	id := ValueId(length - size)
	kv := NewKeyValue(id, key, value)
	if _, err := kv.WriteTo(s.f); err != nil {
		return nil, err
	}
	return kv, nil
}

func (s *FileValueStore) Get(id ValueId) (*KeyValue, error) {
	r := io.NewSectionReader(s.f, int64(id), s.Length()-int64(id))
	var kv KeyValue
	if _, err := kv.ReadFrom(r); err != nil {
		return nil, err
	}
	return &kv, nil
}

func (s *FileValueStore) Each(f func(*KeyValue)) error {
	r := io.NewSectionReader(s.f, 0, s.Length())
	var kv KeyValue
	for _, err := kv.ReadFrom(r); ; _, err = kv.ReadFrom(r) {
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		default:
			f(&kv)
		}
	}
}

func (s *FileValueStore) Sync() error {
	return s.f.Sync()
}

func (s *FileValueStore) Close() error {
	if err := s.f.Sync(); err != nil {
		return err
	}
	return s.f.Close()
}
