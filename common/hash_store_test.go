package jumphasher

import (
	"runtime"
	"sync"
	"testing"
)

func TestNewMemHashStore(t *testing.T) {
	m := NewMemHashStore(runtime.NumCPU())
	if m == nil {
		t.Error("generated MemHashStore should not be nil")
	}
}

func TestMemHashStoreStoreLoad(t *testing.T) {
	type HashPair struct {
		id   UUID
		hash []byte
	}
	hashes := make([]HashPair, 1000000)
	he := NewSHA512Engine()
	hs := NewMemHashStore(runtime.NumCPU())
	wc := make(chan *HashPair)
	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		go func(c chan *HashPair, g *sync.WaitGroup, s HashStore) {
			g.Add(1)
			for x := range c {
				s.Store(&x.id, x.hash)
			}
			g.Done()
		}(wc, &wg, hs)
	}

	for i := 0; i < len(hashes); i++ {
		u, err := UUIDv4()
		if err != nil {
			t.Error(err)
		}
		hashes[i].id = *u
		hashes[i].hash, err = he.Hash(u[:]) //just pretend the password is the Job ID
		if err != nil {
			t.Error(err)
		}
		wc <- &hashes[i] //let the workers add these concurrently
	}
	close(wc)
	wg.Wait()

	//try and recover all the hashes
	for i := 0; i < len(hashes); i++ {
		h, err := hs.Load(&hashes[i].id)
		if err != nil {
			t.Error(err)
		} else if h == nil {
			t.Errorf("Job ID %s stored but could not be located in store", hashes[i].id.MarshalText())
		}
	}

	//try and recover a nonexistent Job ID. should fail
	u2 := hashes[0].id
	u2[0] += 5
	h, err := hs.Load(&u2)
	if err != nil {
		t.Error(err)
	} else if h != nil { //technically this could give a false positive, but probability is extremely low
		t.Errorf("Job ID %s not stored but was located in store", u2.MarshalText())
	}
}
