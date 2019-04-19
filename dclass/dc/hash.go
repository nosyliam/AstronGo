package dc

import "github.com/kavehmz/prime"

const MAX_PRIMES = 10000

type HashGenerator struct {
	primes []uint64
	hash   uint32
	index  uint
}

func NewHashGenerator() *HashGenerator {
	h := &HashGenerator{}
	h.primes = prime.Primes(MAX_PRIMES * 10)
	return h
}

func (g *HashGenerator) AddInt(val int) {
	g.hash += uint32(g.primes[g.index]) * uint32(val)
	g.index = (g.index + 1) % MAX_PRIMES
}

func (g *HashGenerator) AddString(str string) {
	g.AddInt(len(str))
	for _, chr := range str {
		g.AddInt(int(chr))
	}
}

func (g *HashGenerator) Hash() uint32 {
	return g.hash
}
