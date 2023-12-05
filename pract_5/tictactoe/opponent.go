package main

import (
	"math/rand"
	"time"

	"github.com/denilai/maybe"
)

// Сделать случайный шаг, или шаг, который приведет к победе
func HeuristicStep(b Board, me Figure) Place {
	freePlaces := b.FreePlaces()
	for _, p := range freePlaces {
		candidate := Step(b, me, p)
		if candidate.HasValue() && Winner(candidate.FromJust()) == maybe.Just[Figure](me) {
			return p
		}
	}
	rand.Seed(time.Now().UnixNano())
	return freePlaces[rand.Intn(len(freePlaces))]
}
