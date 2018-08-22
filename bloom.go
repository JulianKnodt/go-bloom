package bloom

import (
	"bytes"
	"encoding/binary"
	"sync"
	"sync/atomic"
)

type bloomFilter struct {
	bitfields      map[int][]byte
	bitfieldCounts sync.Map
	totalCount     *uint32
}

type BloomFilter interface {
	Insert(interface{}) error
	PossiblyContains(interface{}) (bool, error)
}

func NewBloomFilter() BloomFilter {
	return &bloomFilter{
		bitfields:      make(map[int][]byte),
		bitfieldCounts: sync.Map{},
		totalCount:     new(uint32),
	}
}

func (b *bloomFilter) Len() uint32 {
	return atomic.LoadUint32(b.totalCount)
}

func (b *bloomFilter) Insert(v interface{}) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return err
	}
	atomic.AddUint32(b.totalCount, 1)
	size := buf.Len()

	parts := buf.Bytes()

	swapped := false
	actual, _ := b.bitfieldCounts.LoadOrStore(size, new(uint32))
	countPointer := actual.(*uint32)
	for !swapped {
		count := atomic.LoadUint32(countPointer)
		bitfield, has := b.bitfields[size]
		if !has {
			bitfield = make([]byte, size)
		}

		for i, v := range bitfield {
			bitfield[i] = v | parts[i]
		}

		b.bitfields[size] = bitfield
		swapped = atomic.CompareAndSwapUint32(countPointer, count, count+1)
	}
	b.bitfieldCounts.Store(size, countPointer)

	bufferPool.Put(buf)
	return nil
}

func (b *bloomFilter) PossiblyContains(v interface{}) (bool, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	buf.Reset()
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		return false, err
	}

	size := buf.Len()
	bitfield, has := b.bitfields[size]
	if !has {
		return false, nil
	}
	parts := buf.Bytes()
	for i, v := range bitfield {
		if v&parts[i] != parts[i] {
			return false, nil
		}
	}
	return true, nil
}
