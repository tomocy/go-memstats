package memstats

import (
	"context"
	"runtime"
)

type Loader interface {
	Load(context.Context) (*runtime.MemStats, error)
}
