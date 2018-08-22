package bloom

import (
	"sync"
	"testing"
)

type Sample struct {
	A int32
	B uint64
}

func TestInsertStruct(t *testing.T) {
	s := Sample{
		A: 10,
		B: 300,
	}

	bf := NewBloomFilter()

	bf.Insert(s)

	if has, err := bf.PossiblyContains(s); err != nil || !has {
		t.Error(err)
	}
}

func BenchmarkInsertFloat64(b *testing.B) {
	bf := NewBloomFilter()
	for i := 0; i < b.N; i++ {
		bf.Insert(float64(1))
	}
}

func TestInsertInt64(t *testing.T) {
	bf := NewBloomFilter()

	if err := bf.Insert(int64(3)); err != nil {
		t.Error(err)
	}

	if has, err := bf.PossiblyContains(int64(3)); err != nil || !has {
		t.Error(err)
	}
}

func TestConcurrentAdd(t *testing.T) {
	bf := NewBloomFilter()
	var wg sync.WaitGroup
	addCount := 100
	wg.Add(1)
	go func() {
		for i := 0; i < addCount; i++ {
			bf.Insert(int64(i))
		}
		wg.Done()
		wg.Wait()

		for i := 0; i < addCount; i++ {
			if has, err := bf.PossiblyContains(int64(i)); !has || err != nil {
				t.Error(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < addCount; i++ {
			bf.Insert(float64(i))
		}
		wg.Done()
		wg.Wait()
		for i := 0; i < addCount; i++ {
			if has, err := bf.PossiblyContains(float64(i)); !has || err != nil {
				t.Error(err)
			}
		}
	}()

	wg.Wait()
}
