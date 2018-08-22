package bloom

import (
	"bytes"
	"sync"
)

var bufferPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
