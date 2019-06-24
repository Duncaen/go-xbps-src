package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Duncaen/xbps-src-go/linter"
)

func main() {
	flag.Parse()
	for _, path := range flag.Args() {
		errs, err := linter.LintFile(path)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Linting: %s\n", path)
		for _, err := range errs {
			fmt.Println(err)
		}
	}
}
