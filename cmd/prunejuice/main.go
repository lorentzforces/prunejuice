package main

import (
	"github.com/lorentzforces/prunejuice/internal/cli"
	"github.com/lorentzforces/prunejuice/internal/run"
)

func main() {
	err := cli.CreateRootCmd().Execute()
	if err == run.ErrorUserDeclined {
		run.CleanFailOut(err.Error())
	}
	run.FailOnErr(err)
}
