package factory

import ()

type Factory[T any] interface {
	Build() T
	Create(out *T) error
	CreateMany(n int) ([]T, error)
}
