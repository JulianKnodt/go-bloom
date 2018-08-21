package bloom

import (
  "testing"
)

func TestInsertStruct(t *testing.T) {

}

func BenchmarkInsertFloat64(b *testing.B) {
  bf := NewBloomFilter()
  for i := 0; i < b.N; i ++ {
    bf.Insert(float64(1))
  }
}
