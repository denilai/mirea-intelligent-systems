package main

func NewMatrix[T any](row, col uint) [][]T {
	a := make([]T, row*col)
	m := make([][]T, row)
	var lo, hi uint = 0, col
	for i := range m {
		m[i] = a[lo:hi:hi]
		lo, hi = hi, hi+col
	}
	return m
}

func TransposeOpt[T any](a [][]T) [][]T {
	b := NewMatrix[T](uint(len(a[0])), uint(len(a)))
	for i := 0; i < len(b); i++ {
		c := b[i]
		for j := 0; j < len(c); j++ {
			c[j] = a[j][i]
		}
	}
	return b
}

func Reduce[T, M any](f func(M, T) M, s []T, initValue M) M {
	acc := initValue
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func Map[T any, M any](f func(T) M, data []T) []M {
	if DEBUG {
		defer Duration(Track("Map"))
	}
	n := make([]M, len(data))
	for i, e := range data {
		n[i] = f(e)
	}
	return n
}

func Filter[T any](f func(T) bool, data []T) []T {
	if DEBUG {
		defer Duration(Track("Filter"))
	}
	fltd := make([]T, 0, len(data))
	for _, e := range data {
		if f(e) {
			fltd = append(fltd, e)
		}
	}
	return fltd
}

func All[T any](f func(T) bool, xs []T) bool {
	return Reduce(func(a bool, b T) bool { return a && f(b) }, xs, true)
}
func Any[T any](f func(T) bool, xs []T) bool {
	return Reduce(func(a bool, b T) bool { return a || f(b) }, xs, true)
}
