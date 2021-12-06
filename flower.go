package florist

import (
	"github.com/hashicorp/go-hclog"
)

type Description struct {
	Name string
	Long string
}

// A flower (made of petals) is what can be installed on a host.
type Flower interface {
	Description() Description
	Install() error
	SetLogger(log hclog.Logger)
}

// A flower just to contain subflowers
type EmptyFlower struct {
	Desc Description
	Log  hclog.Logger
}

func (fl *EmptyFlower) Description() Description {
	return fl.Desc
}

func (fl *EmptyFlower) SetLogger(log hclog.Logger) {
	fl.Log = log
}

func (fl *EmptyFlower) Install() error {
	fl.Log.Info("empty flower install", "desc", fl.Desc)
	return nil
}
