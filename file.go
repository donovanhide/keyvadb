package keyvadb

import (
	"io"
	"os"
)

type FileKeyStore struct {
	f *os.File
}

func NewFileKeyStore(filename string) (KeyStore, error) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &FileKeyStore{
		f: f,
	}, nil
}

func (s *FileKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	return nil, nil
}

func (s *FileKeyStore) Set(node *Node) error {
	return nil
}

func (s *FileKeyStore) Get(id NodeId) (*Node, error) {
	return nil, nil
}

type FileValueStore struct {
	f *os.File
}

func NewFileValueStore(filename string) (ValueStore, error) {
	f, err := os.OpenFile(filename, os.O_APPEND, 0666)
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
	kv := &KeyValue{
		Key: Key{
			Hash: key,
			Id:   ValueId(fi.Size()),
		},
		Value: value,
	}
	if err := kv.MarshalBinary(s.f); err != nil {
		return nil, err
	}
	if err := s.f.Sync(); err != nil {
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
