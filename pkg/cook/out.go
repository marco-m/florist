package cook

import "fmt"

func Out(args ...any) {
	fmt.Print("🥕 ")
	start := 0
	if len(args)%2 == 1 {
		fmt.Printf("%v ", args[0])
		start = 1
	}
	for i := start; i < len(args)-1; i += 2 {
		fmt.Printf("%s=%v ", args[i], args[i+1])
	}
	fmt.Println("")
}
