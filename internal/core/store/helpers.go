package store

import (
	"fmt"
)

const (
	manifestsLatestKey = "latest/"
)

func genLatestKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", manifestsLatestKey, id))
}

func genVersionedKey(id string, version int) []byte {
	return []byte(fmt.Sprintf("%s@v%d", id, version))
}
