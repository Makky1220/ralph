package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
)

var EmbeddedFS fs.FS = embed.FS{}

func BaseFS() (fs.FS, error) {
	return fs.Sub(EmbeddedFS, "templates/base")
}

func PackFS(lang string) (fs.FS, error) {
	packs, err := AvailablePacks()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(packs, lang) {
		return nil, fmt.Errorf("unknown language pack: %q", lang)
	}
	return fs.Sub(EmbeddedFS, "templates/packs/"+lang)
}

func AvailablePacks() ([]string, error) {
	entries, err := fs.ReadDir(EmbeddedFS, "templates/packs")
	if err != nil {
		return nil, err
	}
	var packs []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "_template" {
			packs = append(packs, e.Name())
		}
	}
	return packs, nil
}
