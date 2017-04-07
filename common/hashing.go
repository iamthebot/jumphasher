package jumphasher

import (
	"crypto/sha512"
	"errors"
	"hash"
)

//we could extend this with more hashing algorithms
const (
	HashTypeSHA512 = iota
)

var ErrNilPassword error = errors.New("encountered a nil password")

//Generic hashing interface allows us to swap out hashing algorithms
//Not thread-safe! Use 1 per worker
type HashingEngine interface {
	Hash(password []byte) ([]byte, error)
}

type SHA512Engine struct {
	hasher hash.Hash
}

func NewSHA512Engine() *SHA512Engine {
	var e SHA512Engine
	e.hasher = sha512.New()
	return &e
}

//Generates SHA512 checksum of password
func (e *SHA512Engine) Hash(password []byte) ([]byte, error) {
	if password == nil {
		return nil, ErrNilPassword
	}
	e.hasher.Write(password)
	s := e.hasher.Sum(nil)
	e.hasher.Reset()
	return s, nil
}
