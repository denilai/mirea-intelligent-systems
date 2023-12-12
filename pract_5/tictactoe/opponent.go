package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/denilai/maybe"
)

// Сделать случайный шаг, или шаг, который приведет к победе
func HeuristicStep(b Board, me Figure) (Place, error) {
	freePlaces := b.EmptyPlaces()
	if len(freePlaces) == 0 {
		return Place{}, fmt.Errorf("Нет свободных ходов")
	}
	for _, p := range freePlaces {
		candidate := Step(b, p, me)
		if candidate.HasValue() && Winner(candidate.FromJust()) == maybe.Just[Figure](me) {
			return p, nil
		}
	}
	rand.Seed(time.Now().UnixNano())
	return freePlaces[rand.Intn(len(freePlaces))], nil
}
