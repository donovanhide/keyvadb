package keyvadb

import (
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
	if err := node.UnmarshalBinary(r); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *FileKeyStore) Set(node *Node) error {
	debugPrintln("File Key Set:", node.Id)
	w := ioutil2.NewSectionWriter(s.f, int64(node.Id), NodeBlockSize)
	return node.MarshalBinary(w)
}

type FileValueStore struct {
	f *os.File
}

func NewFileValueStore(filename string) (ValueStore, error) {
	f, err := os.OpenFile(filename+".values", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileValueStore{
		f: f,
	}, nil
}

func (s *FileValueStore) Append(key Hash, value []byte) (*KeyValue, error) {
	fi, err := s.f.Stat()
	if err != nil {
		return nil, err
	}
	kv := NewKeyValue(ValueId(fi.Size()), key, value)
	if err := kv.MarshalBinary(s.f); err != nil {
		return nil, err
	}
	return kv, nil
}

func (s *FileValueStore) Get(id ValueId) (*KeyValue, error) {
	if _, err := s.f.Seek(int64(id), os.SEEK_SET); err != nil {
		return nil, err
	}
	var kv KeyValue
	if err := kv.UnmarshalBinary(s.f); err != nil {
		return nil, err
	}
	return &kv, nil
}

func (s *FileValueStore) Each(f func(*KeyValue)) error {
	r, err := os.Open(s.f.Name())
	if err != nil {
		return err
	}
	var kv KeyValue
	for err = kv.UnmarshalBinary(r); err != nil; err = kv.UnmarshalBinary(r) {
		f(&kv)
	}
	if err == io.EOF {
		return nil
	}
	return err
}
