package jumphasher

import (
	"encoding/binary"
	"errors"
	"sync"
)

var ErrNilJobID error = errors.New("encountered nil job ID")

//we could extend this with alternative hash stores
const (
	HashStoreTypeMem = iota
)

//Implements a generic way to store and load job IDs and their corresponding hashes
//Lets us easily swap out backends, eg; in-memory, RDBMS like Postgres, NoSQL store
//Underlying type *must be thread safe*
type HashStore interface {
	Store(id *UUID, h []byte) error
	Load(id *UUID) ([]byte, error)
}

//In-memory hash store using granular locking to prevent contention
//
//Items are routed to a bucket based on first 4 bytes of job ID
//
//Ideally, we size this to the hardware concurrency of the machine
type MemHashStore struct {
	buckets []map[UUID][]byte
	locks   []sync.Mutex
}

//Creates a new MemHashStore object
func NewMemHashStore(size int) *MemHashStore {
	var m MemHashStore
	m.buckets = make([]map[UUID][]byte, size)
	for i := range m.buckets {
		m.buckets[i] = make(map[UUID][]byte)
	}
	m.locks = make([]sync.Mutex, size)
	return &m
}

//Store a hash given a job ID
func (m *MemHashStore) Store(id *UUID, h []byte) error {
	if id == nil {
		return ErrNilJobID
	}
	//calculate bucket
	head := binary.LittleEndian.Uint32(id[0:4])
	slot := int(head) % len(m.buckets)

	//lock the corresponding bucket
	m.locks[slot].Lock()
	//store the item
	m.buckets[slot][*id] = h
	//unlock the corresponding bucket
	m.locks[slot].Unlock()
	return nil
}

//Load a hash given a job ID
//
//If it cannot be found, we return a nil slice
func (m *MemHashStore) Load(id *UUID) ([]byte, error) {
	if id == nil {
		return nil, ErrNilJobID
	}

	//calculate bucket
	head := binary.LittleEndian.Uint32(id[0:4])
	slot := int(head) % len(m.buckets)

	//lock the corresponding bucket
	m.locks[slot].Lock()
	//load the item
	h, exists := m.buckets[slot][*id]
	//unlock the corresponding bucket
	m.locks[slot].Unlock()

	if !exists {
		return nil, nil
	}
	return h, nil
}
