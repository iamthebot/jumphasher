package jumphasher

//Contains a super primitive UUIDv4 generator and string encoder/decoder
//Used for generating unique Job IDs
import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

type UUID [16]byte //use a fixed size array so it's hashable (i.e. usable as map[UUID]interface{})

// Generates an RFC 4122 compliant UUIDv4.
//
// Returns a 128 bit byte slice containing the UUID
//
// Note: uses crypto/rand for secure entropy so it will be slow if the underlying CPU doesn't support the RDRAND instruction.
// Since we only need low collision probability and even distribution,
// hashing the output of a faster PRNG (eg; mersenne-twister) through something like 64 bit xxhash twice (to get 16 bytes)
// will be considerably less CPU intensive.
func UUIDv4() (*UUID, error) {
	//temporary buffer for 16 bytes using crypto PRNG
	var u UUID
	_, err := rand.Read(u[:])
	if err != nil {
		return nil, err
	}
	//set highest 4 bits of time_hi_and_version (octets 6-7) to 4 (for UUID v4)
	//see: http://pubs.opengroup.org/onlinepubs/9629399/apdxa.htm
	u[6] = (u[6] & 0x0F) | (byte(4) << 4)
	//sets variant in octet 8
	u[8] = (u[8] | 0x40) & 0x7f
	return &u, nil
}

//Outputs a UUID in unhyphenated hex
//
//Technically, this isn't the standard format but it's easier to deal with than hyphenated variants
//
//Implements encoding.TextMarshaler interface which allows easy usage with JSON
func (u *UUID) MarshalText() string {
	return hex.EncodeToString(u[:])
}

//Parses a UUID from string using unhyphenated hex format.
//
//Input is expected to be a lowercase 32 character alphanumeric string
//
//Implements encoding.TextUnmarshaler interface which allows easy usage with JSON
func (u *UUID) UnmarshalText(s string) error {
	if len(s) != 32 {
		return errors.New("UUID string must be in unhyphenated hex format")
	}
	//validate input first to make sure it's sane
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) { //check we only use lowercase alphanumeric characters
			return fmt.Errorf("Could not parse UUID: encountered invalid character %0x", c)
		}
	}
	buf, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("Could not decode UUID: %s", err.Error())
	}
	copy(u[:], buf)
	return nil
}

//Convenience function to compare two UUIDs
//
//Will return false if either input is nil
func (a *UUID) Equals(b *UUID) bool {
	if a == nil || b == nil {
		return false
	}
	return bytes.Equal(a[:], b[:])
}
