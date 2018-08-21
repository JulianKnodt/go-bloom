package bloom

import (
	"bytes"
	"encoding/gob"
	"sync/atomic"
)

type bloomFilter struct {
	bitfields      map[int][]byte
	bitfieldCounts map[int]*int
	totalCount     *uint32
}

type BloomFilter interface {
	Insert(interface{})
	PossiblyContains(interface{}) bool
}

func NewBloomFilter() BloomFilter {
	return &bloomFilter{
		bitfields:      make(map[int][]byte),
		bitfieldCounts: make(map[int]*int),
		totalCount:     new(uint32),
	}
}

func (b bloomFilter) Len() uint32 {
	return atomic.LoadUint32(b.totalCount)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func (b *bloomFilter) Insert(v interface{}) {
	buf := bufferPool.Get().(*bytes.Buffer)
	Must(gob.NewEncoder(buf).Encode(v))

	size := buf.Len()
	bitfield, has := b.bitfields[size]
	if !has {
		bitfield = make([]byte, size)
		b.bitfields[size] = bitfield
	}
	atomic.AddUint32(b.totalCount, 1)

	parts := buf.Bytes()
	for i, v := range bitfield {
		bitfield[i] = v | parts[i]
	}

	b.bitfields[size] = bitfield

  buf.Reset()
  bufferPool.Put(buf)
}

func (b bloomFilter) PossiblyContains(v interface{}) bool {
	buf := new(bytes.Buffer)
	Must(gob.NewEncoder(buf).Encode(v))

	size := buf.Len()
	bitfield, has := b.bitfields[size]
	if !has {
		return false
	}

	parts := buf.Bytes()
	for i, v := range bitfield {
		if v&parts[i] != parts[i] {
			return false
		}
	}
	return true
}
