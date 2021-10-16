package florist

// A flower (made of petals) is what can be installed on a host.
type Flower interface {
	Install() error
	String() string
}
