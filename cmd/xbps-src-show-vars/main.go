package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Duncaen/go-xbps-src/template"
)

var (
	machine = flag.String("machine", "x86_64", "architecture")
	targetMachine = flag.String("target-machine", "", "target architecture")
)

func main() {
	flag.Parse()
	env := template.Environment(*machine, *targetMachine)
	for _, path := range flag.Args() {
		t, err := template.ParseFile(path)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Evaluating: %s\n", path)
		vars, err := t.Eval(env)
		for k, v := range vars {
			fmt.Printf("%s=%q\n", k, v)
		}
	}
}
