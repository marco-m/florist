// Package Salt provides helpers when migrating from Salt to Florist.
package salt

import (
	"runtime"
	"strings"
)

// Platform returns runtime.GOOS, but with the first letter uppercase.
// For example: "linux" -> "Linux".
// This is close (but maybe not always) to what Salt grains.kernel would return.
func Platform() string {
	// Although strings.Title is deprecated, it works well with runtime.GOOS,
	// which I assume never has whitespace it it.
	//lint:ignore SA1019 it works well with runtime.GOOS
	return strings.Title(runtime.GOOS)
}
