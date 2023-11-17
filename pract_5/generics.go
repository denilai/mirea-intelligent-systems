package main

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
