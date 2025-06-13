package main

import (
	"fmt"
	"os"
)

func execute(args []string) {
	fmt.Println("execute called with args:", args)
}

func main() {
	if len(os.Args) > 1 {
		execute(os.Args[1:])
	} else {
		fmt.Println("No CLI input provided.")
	}
}
