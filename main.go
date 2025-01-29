package main

import (
	"log"
	"os"

	"github.com/javiorfo/passcualito/passc"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[ERROR] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	passc.Builder().Execute()
}
