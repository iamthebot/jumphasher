package jumphasher

import (
	"testing"
)

func TestUUIDv4(t *testing.T) {
	u, err := UUIDv4()
	if err != nil {
		t.Error(err)
	} else if u == nil {
		t.Error("Expected non-nil UUID ptr")
	}
	//Check that we actually got a version 4 UUID
	if uint8(u[6]>>4) != uint8(4) {
		t.Errorf("UUID Version. Recieved: %#x Expected: %#x", uint8(u[6]>>4), uint8(4))
	}

	if uint8(u[8]&0x40) != uint8(0x40) {
		t.Errorf("UUID Variant. Recieved: %#x Expected: %#x", uint8(u[8]&0x40), uint8(0x40))
	}

	//Do a quick sanity check and make sure there are no collisions in 1000 iterations
	m := make(map[UUID]byte)
	for i := 0; i < 1000; i++ {
		u, err := UUIDv4()
		if err != nil {
			t.Error(err)
		}
		_, exists := m[*u]
		if exists {
			t.Errorf("UUID Collision on iteration %d", i)
		}
		m[*u] = 0x1
	}
}

func BenchmarkUUIDv4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := UUIDv4()
		if err != nil {
			b.Error(err)
		}
	}
}

func TestUUID_MarshalText(t *testing.T) {
	u, err := UUIDv4()
	if err != nil {
		t.Error(err)
	} else if u == nil {
		t.Error("Expected non-nil UUIDv4")
	}
	s := u.MarshalText()

	//check length
	if len(s) < 1 {
		t.Errorf("Output is of invalid length: %d", len(s))
	}
}

func BenchmarkUUID_MarshalText(b *testing.B) {
	for i := 0; i < b.N; i++ {
		u, err := UUIDv4()
		if err != nil {
			b.Error(err)
		}
		u.MarshalText()
	}
}

func TestUUID_UnmarshalText(t *testing.T) {
	u, err := UUIDv4()
	if err != nil {
		t.Error(err)
	} else if u == nil {
		t.Error("Expected non-nil UUIDv4")
	}
	s := u.MarshalText()
	//now try decoding it
	var u2 UUID
	err = u2.UnmarshalText(s)
	if err != nil {
		t.Error(err)
	}
	if !u.Equals(&u2) {
		t.Errorf("Decoded UUID doesn't match. Expected: %s Recieved: %s", s, u2.MarshalText())
	}
}

//Technically, the unmarshaling time is roughly half of this benchmark
func BenchmarkUUID_UnmarshalText(b *testing.B) {
	for i := 0; i < b.N; i++ {
		u, err := UUIDv4()
		if err != nil {
			b.Error(err)
		}
		s := u.MarshalText()
		var u2 UUID
		u2.UnmarshalText(s)
	}
}

func TestUUID_Equals(t *testing.T) {
	u, err := UUIDv4()
	if err != nil {
		t.Error(err)
	} else if u == nil {
		t.Error("Expected non-nil UUID ptr")
	}
	if !u.Equals(u) {
		t.Error("UUID must equal itself")
	}
	var u2 UUID = *u
	u2[5] += 1
	if u.Equals(&u2) {
		t.Error("Differing UUIDs must not be equal")
	}
}

func BenchmarkUUID_Equals(b *testing.B) {
	for i := 0; i < b.N; i++ {
		u, err := UUIDv4()
		if err != nil {
			b.Error(err)
		}
		var u2 UUID
		u2 = *u
		u.Equals(&u2)
	}
}
