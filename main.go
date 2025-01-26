package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/danp/scraperlite/internal"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func main() {
	if err := internal.Run(os.Args, os.Stdout, time.Now); errors.Is(err, internal.ErrExit1) {
		os.Exit(1)
	} else if err != nil {
		log.Fatal(err)
	}
}
