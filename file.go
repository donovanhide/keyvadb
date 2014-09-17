package keyvadb

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type FileStoreConfig struct {
	LastId ValueId
	Degree uint64
}

type FileKeyStore struct {
	FileStoreConfig
	f *os.File
}

var rootNodeId = NodeId(binary.Size(FileStoreConfig{}))

func NewFileKeyStore(degree uint64, filename string) (KeyStore, error) {
	f, err := os.OpenFile(filename+".keys", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	ks := &FileKeyStore{f: f}
	if err := binary.Read(ks.f, binary.BigEndian, ks.FileStoreConfig); err != nil {
		ks.Degree = degree
		if err := binary.Write(ks.f, binary.BigEndian, ks.FileStoreConfig); err != nil {
			return nil, err
		}
		if err := f.Sync(); err != nil {
			return nil, err
		}
	} else {
		if ks.Degree != degree {
			return nil, fmt.Errorf("Cannot use file with different degree: %d file: %d", degree, ks.Degree)
		}
	}
	return ks, nil
}

func (s *FileKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	debugPrintln("File New:", start, end, degree)
	n, err := s.f.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	node := NewNode(start, end, NodeId(n), degree)
	if err := node.MarshalBinary(s.f); err != nil {
		return nil, err
	}
	if err := s.f.Sync(); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *FileKeyStore) Get(id NodeId, degree uint64) (*Node, error) {
	debugPrintln("File Get:", id, degree)
	if _, err := s.f.Seek(int64(id), os.SEEK_SET); err != nil {
		return nil, err
	}
	node := NewNode(FirstHash, LastHash, id, degree)
	err := node.UnmarshalBinary(s.f)
	switch {
	case err == io.EOF:
		return nil, ErrNotFound
	case err != nil:
		return nil, err
	default:
		return node, nil
	}
}

func (s *FileKeyStore) Set(node *Node) error {
	debugPrintln("File Set:", node.Id)
	if _, err := s.f.Seek(int64(node.Id), os.SEEK_SET); err != nil {
		return err
	}
	if err := node.MarshalBinary(s.f); err != nil {
		return err
	}
	return s.f.Sync()
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
