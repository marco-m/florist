package florist

import "github.com/hashicorp/go-hclog"

// A flower (made of petals) is what can be installed on a host.
type Flower interface {
	Install() error
	SetLogger(log hclog.Logger)
	String() string
}
