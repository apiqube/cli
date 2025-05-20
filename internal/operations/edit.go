package operations

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/apiqube/cli/internal/core/manifests"
	"gopkg.in/yaml.v3"
)

var ErrFileNotEdited = errors.New("file was not edited")

func EditFormat(format ParseFormat, manifest manifests.Manifest) (manifests.Manifest, error) {
	if format == JSONFormat {
		return editAsJson(manifest)
	} else {
		return editAsYaml(manifest)
	}
}

func Edit(manifest manifests.Manifest) (manifests.Manifest, error) {
	return editAsYaml(manifest)
}

func editAsYaml(manifest manifests.Manifest) (manifests.Manifest, error) {
	tmpFile, _ := os.CreateTemp("", fmt.Sprintf("%s.*.yaml", manifest.GetID()))
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	var data []byte
	var err error

	if data, err = yaml.Marshal(manifest); err != nil {
		return manifest, fmt.Errorf("error marshalling manifest: %s", err.Error())
	}

	if _, err = tmpFile.Write(data); err != nil {
		return manifest, fmt.Errorf("error writing manifest data to temp file: %s", err.Error())
	}

	if err = tmpFile.Close(); err != nil {
		return manifest, fmt.Errorf("error closing temp file: %s", err.Error())
	}

	if err = editManifestFile(tmpFile.Name()); err != nil {
		if errors.Is(err, ErrFileNotEdited) {
			return manifest, err
		}

		return manifest, fmt.Errorf("error editing manifest: %s", err.Error())
	}

	var updatedData []byte
	if updatedData, err = os.ReadFile(tmpFile.Name()); err != nil {
		return manifest, fmt.Errorf("error reading updated manifest: %s", err.Error())
	}

	var result manifests.Manifest

	if result, err = Parse(YAMLFormat, updatedData); err != nil {
		return manifest, err
	}

	return result, nil
}

func editAsJson(manifest manifests.Manifest) (manifests.Manifest, error) {
	tmpFile, _ := os.CreateTemp("", fmt.Sprintf("%s.*.json", manifest.GetID()))
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	var data []byte
	var err error

	if data, err = json.Marshal(manifest); err != nil {
		return manifest, fmt.Errorf("error marshalling manifest: %s", err.Error())
	}

	if _, err = tmpFile.Write(data); err != nil {
		return manifest, fmt.Errorf("error writing manifest data to temp file: %s", err.Error())
	}

	if err = tmpFile.Close(); err != nil {
		return manifest, fmt.Errorf("error closing temp file: %s", err.Error())
	}

	if err = editManifestFile(tmpFile.Name()); err != nil {
		if errors.Is(err, ErrFileNotEdited) {
			return manifest, err
		}

		return manifest, fmt.Errorf("error editing manifest: %s", err.Error())
	}

	var updatedData []byte
	if updatedData, err = os.ReadFile(tmpFile.Name()); err != nil {
		return manifest, fmt.Errorf("error reading updated manifest: %s", err.Error())
	}

	var result manifests.Manifest

	if result, err = Parse(JSONFormat, updatedData); err != nil {
		return manifest, err
	}

	return result, nil
}

func editManifestFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access file: %w", err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	newInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to verify edited file: %w", err)
	}

	if fileInfo.ModTime().Equal(newInfo.ModTime()) {
		return ErrFileNotEdited
	}

	return nil
}
