package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func CalculateHashWithPath(filePath string, content []byte) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	modTime := fileInfo.ModTime().UnixNano()

	hasher := sha256.New()
	hasher.Write(content)
	hasher.Write([]byte(fmt.Sprintf("%d", modTime)))

	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, nil
}
