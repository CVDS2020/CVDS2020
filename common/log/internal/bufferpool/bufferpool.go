package bufferpool

import (
	"github.com/CVDS2020/CVDS2020/common/pool"
)

var (
	_pool = pool.NewBufferPool(1024)
	// Get retrieves a buffer from the pool, creating one if necessary.
	Get = _pool.Get
)
