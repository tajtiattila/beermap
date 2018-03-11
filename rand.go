package main

import (
	crand "crypto/rand"
	"io"
	"math/big"
	"strings"
)

// alphabet from Bitcoin address and IPFS hash
const btcAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var mapKeyGen = idGen(crand.Reader, btcAlphabet, 12)
var editPassGen = idGen(crand.Reader, btcAlphabet, 16)

type idGenerator struct {
	reader   io.Reader
	alphabet string
	length   int
}

func idGen(r io.Reader, alphabet string, length int) idGenerator {
	return idGenerator{
		reader:   r,
		alphabet: alphabet,
		length:   length,
	}
}

func (g idGenerator) generateKey() (string, error) {
	if g.length <= 0 || len(g.alphabet) == 0 {
		return "", nil
	}
	v := big.NewInt(1)
	base := big.NewInt(int64(len(g.alphabet)))
	for i := 0; i < g.length; i++ {
		v.Mul(v, base)
	}

	nbytes := (v.BitLen() + 7) / 8
	buf := make([]byte, nbytes)

	_, err := io.ReadFull(g.reader, buf)
	if err != nil {
		return "", err
	}

	v.SetBytes(buf)

	s := make([]byte, g.length)
	m := big.NewInt(0)
	for i := range s {
		v.DivMod(v, base, m)
		s[i] = g.alphabet[int(m.Int64())]
	}

	return string(s), nil
}

func (g idGenerator) validKey(s string) bool {
	if len(s) != g.length {
		return false
	}
	for _, r := range s {
		if strings.IndexRune(g.alphabet, r) == -1 {
			return false
		}
	}
	return true
}
