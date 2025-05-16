package main

import (
	"github.com/apiqube/cli/cmd/cli"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/ui"
	"time"
)

func main() {
	ui.Init()
	store.Init()
	defer store.Stop()
	cli.Execute()
	ui.StopWithTimeout(time.Second)
}
