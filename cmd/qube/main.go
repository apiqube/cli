package main

import (
	"github.com/apiqube/cli/internal/core/store"
	uicli "github.com/apiqube/cli/ui/cli"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

func main() {
	uicli.Init()
	defer uicli.Stop()

	store.Init()
	defer store.Stop()

	Execute()
}
