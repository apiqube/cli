package main

import (
	uicli "github.com/apiqube/cli/ui/cli"
	"time"

	"github.com/apiqube/cli/cmd/cli"
	"github.com/apiqube/cli/internal/core/store"
)

func main() {
	uicli.Init()
	defer uicli.Stop()

	store.Init()
	defer store.Stop()

	cli.Execute()

	time.Sleep(time.Second)
}
