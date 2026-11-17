package xpi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
)

type Gecko struct {
	Id        *string `json:"id"`
	UpdateUrl *string `json:"update_url"`
}

type BrowserSpecificSettings struct {
	Gecko *Gecko `json:"gecko"`
}

type Manifest struct {
	// ManifestVersion *string `json:"manifest_version"`
	Version                 string                   `json:"version"`
	Description             string                   `json:"description"`
	Name                    string                   `json:"name"`
	BrowserSpecificSettings *BrowserSpecificSettings `json:"browser_specific_settings"`
}

func ReadZip(data []byte) (*zip.Reader, error) {
	reader := bytes.NewReader(data)

	return zip.NewReader(reader, int64(len(data)))
}

func ExtractManifestFromArchive(data []byte) (*Manifest, error) {
	archive, err := ReadZip(data)
	if err != nil {
		return nil, err
	}
	for _, f := range archive.File {
		if f.Name != "manifest.json" {
			continue
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return nil, err
		}
		log.Println("Found manifest.json in archive")
		content, err := io.ReadAll(fileInArchive)
		if err != nil {
			return nil, err
		}

		var manifest Manifest
		err = json.Unmarshal(content, &manifest)
		if err != nil {
			return nil, err
		}

		fileInArchive.Close()
		return &manifest, nil
	}
	return nil, os.ErrNotExist
}
