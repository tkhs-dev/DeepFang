package rule

import "math/rand"

type Rand interface {
	Intn(n int) int
	// Range produces a random integer in the range of [min,max] (inclusive)
	// For example, Range(1,3) could produce the values 1, 2 or 3.
	// Panics if max < min (like how Intn(n) panics for n <=0)
	Range(min, max int) int
	Shuffle(n int, swap func(i, j int))
}

type StandardRand struct {
	seed int64
	rand *rand.Rand
}

func NewStandardRand(seed int64) *StandardRand {
	return &StandardRand{
		seed: seed,
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (s StandardRand) Intn(n int) int {
	return s.rand.Intn(n)
}

func (s StandardRand) Range(min, max int) int {
	return s.rand.Intn(max-min+1) + min
}

func (s StandardRand) Shuffle(n int, swap func(i, j int)) {
	s.rand.Shuffle(n, swap)
}
