package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Must supply 1 parameter\n")
		os.Exit(1)
	}
	id10s := os.Args[1]
	id10_64, err := strconv.ParseInt(id10s, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid string ->%s<- for base 10: %s\n", id10s, err)
		os.Exit(1)
	}
	id10 := int(id10_64)
	id36 := strconv.FormatUint(uint64(id10), 36) // Base 36, take count of # of files add 1, this is the code.
	fmt.Printf("%s\n", id36)
}
