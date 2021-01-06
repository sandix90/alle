package main

import (
	"alle/internal/cmd"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	commander, err := cmd.NewCommander()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = commander.Execute()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}
