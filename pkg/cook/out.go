package cook

import (
	"fmt"
	"strings"
)

func Out(args ...any) {
	var bld strings.Builder
	fmt.Fprint(&bld, "🥕 ")
	start := 0
	if len(args)%2 == 1 {
		fmt.Fprintf(&bld, "%v ", args[0])
		start = 1
	}
	for i := start; i < len(args)-1; i += 2 {
		fmt.Fprintf(&bld, "%s=%v ", args[i], args[i+1])
	}
	fmt.Fprintln(&bld, "")

	fmt.Print(bld.String())
}
