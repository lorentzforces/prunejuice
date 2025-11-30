package main

import (
	"github.com/lorentzforces/prunejuice/internal/cli"
	"github.com/lorentzforces/prunejuice/internal/run"
)

func main() {
	err := cli.CreateRootCmd().Execute()
	run.FailOnErr(err)
}
