package store

import "fmt"

func genManifestKey(id string) []byte {
	return []byte(id)
}

func genManifestListKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", manifestListKeyPrefix, id))
}

func genManifestHashKey(hash string) []byte {
	return []byte(fmt.Sprintf("%s%s", manifestHashKeyPrefix, hash))
}
