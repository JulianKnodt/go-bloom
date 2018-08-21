package bloom

import (
  "sync"
  "bytes"
)

var bufferPool sync.Pool = sync.Pool{
  New: func() interface{} {
    return new(bytes.Buffer)
  },
}


