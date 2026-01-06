package factory

import (
	"sync/atomic"
	"time"

	"github.com/galaplate/core/database"
	"gorm.io/gorm"
)

type BaseFactory[T any] struct {
	db      *gorm.DB
	seq     *int64
	Builder func(seq int64) T
}

// NewBaseFactory creates a new BaseFactory for a model
func NewBaseFactory[T any](builder func(seq int64) T) *BaseFactory[T] {
	var s int64 = 0
	return &BaseFactory[T]{db: database.Connect, seq: &s, Builder: builder}
}

func (f *BaseFactory[T]) nextSeq() int64 {
	return atomic.AddInt64(f.seq, 1)
}

// Build creates an instance but does not persist
func (f *BaseFactory[T]) Build() T {
	seq := f.nextSeq()
	return f.Builder(seq)
}

// Create builds and persists a new record
func (f *BaseFactory[T]) Create(out *T) error {
	val := f.Build()
	if err := f.db.Create(&val).Error; err != nil {
		return err
	}
	*out = val
	return nil
}

// CreateMany builds and persists multiple records
func (f *BaseFactory[T]) CreateMany(n int) ([]T, error) {
	result := make([]T, 0, n)
	for range n {
		val := f.Build()
		if err := f.db.Create(&val).Error; err != nil {
			return nil, err
		}
		result = append(result, val)
		time.Sleep(1 * time.Millisecond)
	}
	return result, nil
}
