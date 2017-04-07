package jumphasher

import (
	"encoding/hex"
	"testing"
)

func TestNewSHA512Engine(t *testing.T) {
	e := NewSHA512Engine()
	if e == nil {
		t.Error("New SHA512 engine must not be nil")
	}
}

//tautological, but useful for regression testing
func TestSHA512EngineHash(t *testing.T) {
	vectors := make([][]byte, 5) //contains the NIST-approved test vectors
	vectors[0] = []byte("abc")
	vectors[1] = []byte("")
	vectors[2] = []byte("abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq")
	vectors[3] = []byte("abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu")
	vectors[4] = make([]byte, 1000000) //1 million a's
	for i := 0; i < len(vectors[4]); i++ {
		vectors[4][i] = 'a'
	}
	hashes := make([]string, 5) //contains verified hex digests corresponding to NIST test vectors
	hashes[0] = "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"
	hashes[1] = "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	hashes[2] = "204a8fc6dda82f0a0ced7beb8e08a41657c16ef468b228a8279be331a703c33596fd15c13b1b07f9aa1d3bea57789ca031ad85c7a71dd70354ec631238ca3445"
	hashes[3] = "8e959b75dae313da8cf4f72814fc143f8f7779c6eb9f7fa17299aeadb6889018501d289e4900f7e4331b99dec4b5433ac7d329eeb6dd26545e96e55b874be909"
	hashes[4] = "e718483d0ce769644e2e42c7bc15b4638e1f98b13b2044285632a803afa973ebde0ff244877ea60a4cb0432ce577c31beb009c5c2c49aa2e4eadb217ad8cc09b"

	he := NewSHA512Engine()
	for i := 0; i < len(vectors); i++ {
		h, err := he.Hash(vectors[i])
		if err != nil {
			t.Error(err)
		} else if hex.EncodeToString(h) != hashes[i] {
			t.Errorf("Expected: %s Actual: %s", hashes[i], hex.EncodeToString(h))
		}
	}
}
