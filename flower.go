package florist

import "fmt"

// A flower is a composable unit that can be installed.
type Flower interface {
	fmt.Stringer
	Description() string
	Install() error
}
