package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
)

type Figure struct {
	Phrase string
	Font   string
	Color  string
	Strict bool
}

func (f *Figure) PreHandle() config.PreHandlerConfig {
	if f == nil {
		f = new(Figure)
	}
	f.Phrase = "CVDSMSU"
	return f
}
