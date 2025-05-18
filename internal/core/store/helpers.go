package store

import (
	"fmt"
	"strings"
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

func extractBaseID(versionedKey string) string {
	if at := strings.LastIndex(versionedKey, "@"); at != -1 {
		return versionedKey[:at]
	}
	return versionedKey
}
