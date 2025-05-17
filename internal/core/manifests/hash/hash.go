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
	if _, err = fmt.Fprintf(hasher, "%d", modTime); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %s", err.Error())
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, nil
}

func CalculateHashWithContent(content []byte) (string, error) {
	hasher := sha256.New()
	hasher.Write(content)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, nil
}
