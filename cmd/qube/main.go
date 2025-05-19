package main

import (
	"time"

	"github.com/apiqube/cli/cmd/cli"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
)

func main() {
	ui.Init()
	defer ui.Stop()

	store.Init()
	defer store.Stop()

	cli.Execute()

	time.Sleep(time.Second)
}
