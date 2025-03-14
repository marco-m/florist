package internal

import "fmt"

func MakeErrorf(prefix string) func(format string, a ...any) error {
	return func(format string, a ...any) error {
		return fmt.Errorf(prefix+": "+format, a...)
	}
}
