package units

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/apiqube/cli/internal/core/io"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const testDataPath = "testdata"

func TestCoreIOLoad(t *testing.T) {
	store.InitWithPath(testDataPath)
	defer func() {
		store.Stop()

		storePath, err := xdg.DataFile(testDataPath)
		require.NoError(t, err)

		err = os.RemoveAll(storePath)
		require.NoError(t, err)
	}()

	t.Run("TestCoreIOLoad: nothing to load", func(t *testing.T) {
		newMans, cachedMans, _ := io.LoadManifests(".")

		require.Len(t, newMans, 0)
		require.Len(t, cachedMans, 0)
	})

	t.Run("TestCoreIOLoad: load server manifest", func(t *testing.T) {
		newMans, cachedMans, err := io.LoadManifests(fmt.Sprintf("%s/test_server.yaml", testDataPath))

		require.NoError(t, err)
		require.Len(t, newMans, 1)
		require.Len(t, cachedMans, 0)
		require.Equal(t, manifests.ServerKind, newMans[0].GetKind())
	})

	t.Run("TestCoreIOLoad: load several manifests", func(t *testing.T) {
		newMans, cachedMans, err := io.LoadManifests(testDataPath)

		require.NoError(t, err)
		require.Len(t, newMans, 2)
		require.Len(t, cachedMans, 0)
	})
}
