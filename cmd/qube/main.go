package main

import (
	"github.com/apiqube/cli/cmd/cli"
	"github.com/apiqube/cli/internal/core/store"
	uicli "github.com/apiqube/cli/ui/cli"
)

func main() {
	uicli.Init()
	defer uicli.Stop()

	store.Init()
	defer store.Stop()

	cli.Execute()
}
