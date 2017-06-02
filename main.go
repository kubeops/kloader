package main

import (
	"os"

	"github.com/appscode/log"
)

func main() {
	cmd := newKloaderCmd()
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
