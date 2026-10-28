package main

import (
	"encoding/json"
	"io/fs"
)

// suiteAssets keeps original per-game URLs while embedding shared media once.
type suiteAssets struct {
	base    fs.FS
	aliases map[string]string
}

func newSuiteAssets(base fs.FS) fs.FS {
	data, err := fs.ReadFile(base, "asset-map.json")
	if err != nil {
		panic(err)
	}
	var aliases map[string]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		panic(err)
	}
	return &suiteAssets{base: base, aliases: aliases}
}

func (s *suiteAssets) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if target, ok := s.aliases[name]; ok {
		name = target
	}
	return s.base.Open(name)
}
