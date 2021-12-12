package florist

// A flower is a composable unit that can be installed.
type Flower interface {
	Name() string
	Description() string
	Install() error
}
