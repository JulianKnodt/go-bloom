package bloom

import (
	"bytes"
	"encoding/binary"
	"sync"
	"sync/atomic"
)

type bloomFilter struct {
	bitfields     map[int][]byte
	bitfieldLocks sync.Map // map[int]*sync.RWMutex
	totalCount    *uint32
}

type BloomFilter interface {
	Insert(interface{}) error
	PossiblyContains(interface{}) (bool, error)
	Len() uint32
}

func NewBloomFilter() BloomFilter {
	return &bloomFilter{
		bitfields:     make(map[int][]byte),
		bitfieldLocks: sync.Map{},
		totalCount:    new(uint32),
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

	val, has := b.bitfieldLocks.Load(size)
	if !has {
		val, _ = b.bitfieldLocks.LoadOrStore(size, &sync.RWMutex{})
	}
	mutex := val.(*sync.RWMutex)
	mutex.Lock()

	bitfield, has := b.bitfields[size]
	if !has {
		bitfield = make([]byte, size)
		b.bitfields[size] = bitfield
	}

	for i := range bitfield {
		bitfield[i] |= parts[i]
	}

	b.bitfields[size] = bitfield
	mutex.Unlock()

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
	parts := buf.Bytes()

	val, has := b.bitfieldLocks.Load(size)
	if !has {
		return false, nil
	}
	mutex := val.(*sync.RWMutex)

	mutex.RLock()
	defer mutex.RUnlock()

	bitfield, has := b.bitfields[size]
	if !has {
		return false, nil
	}

	for i, v := range bitfield {
		if v&parts[i] != parts[i] {
			return false, nil
		}
	}
	return true, nil
}
