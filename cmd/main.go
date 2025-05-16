package main

import (
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/internal/ui"
	"github.com/dgraph-io/badger/v4/badger/cmd"
	"time"
)

func main() {
	ui.Init()
	defer ui.StopWithTimeout(time.Microsecond * 250)

	store.Init()
	defer store.Stop()

	cmd.Execute()
}
