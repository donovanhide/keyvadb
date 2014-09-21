package keyvadb

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/siddontang/go/ioutil2"
)

type FileKeyStore struct {
	f      *os.File
	length int64
}

func NewFileKeyStore(degree uint64, filename string) (KeyStore, error) {
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
	return &FileKeyStore{f: f, length: fi.Size()}, nil
}

func (s *FileKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	offset := atomic.AddInt64(&s.length, NodeBlockSize)
	debugPrintln("File Key New:", offset)
	node := NewNode(start, end, NodeId(offset), degree)
	return node, nil
}

func (s *FileKeyStore) Get(id NodeId, degree uint64) (*Node, error) {
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

type FileValueStore struct {
	f      *os.File
	length int64
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

func (s *FileValueStore) Append(key Hash, value []byte) (*KeyValue, error) {
	length := int64(SizeOfKeyValue(value))
	id := ValueId(atomic.AddInt64(&s.length, length) - length)
	kv := NewKeyValue(id, key, value)
	w := bufio.NewWriter(s.f)
	if err := kv.MarshalBinary(w); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return kv, nil
}

func (s *FileValueStore) Get(id ValueId) (*KeyValue, error) {
	length := atomic.LoadInt64(&s.length)
	r := io.NewSectionReader(s.f, int64(id), length)
	var kv KeyValue
	if err := kv.UnmarshalBinary(r); err != nil {
		return nil, err
	}
	return &kv, nil
}

func (s *FileValueStore) Each(f func(*KeyValue)) error {
	length := atomic.LoadInt64(&s.length)
	r := io.NewSectionReader(s.f, 0, length)
	var kv KeyValue
	for err := kv.UnmarshalBinary(r); ; err = kv.UnmarshalBinary(r) {
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
