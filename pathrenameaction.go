package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func NewPathRenameAction() Action {
	return &PathRenameAction{
		Excludes:      make([]string, 0),
		Substitutions: make([]Substitution, 0),
	}
}

type PathRenameAction struct {
	Dir           string
	Excludes      []string
	Substitutions []Substitution
}

func (a *PathRenameAction) Set(s ActionSpec) error {
	return s.ToAction(a)
}

func (a *PathRenameAction) Validate() error {
	return nil
}

func (a *PathRenameAction) processDir(path string) error {
	log.Println("processing dir: ", path)
	for _, s := range a.Substitutions {
		newPath := strings.ReplaceAll(path, s.Old, s.New)
		if path != newPath {
			err := os.Rename(path, newPath)
			if err != nil {
				return err
			}
			// this limits the operation to a depth of 1, but otherwise
			return filepath.SkipDir
		}
	}
	return nil
}

func (a *PathRenameAction) skipExcluded(name string) error {
	for _, excl := range a.Excludes {
		if name == excl {
			return filepath.SkipDir
		}
	}
	return nil
}

func (a *PathRenameAction) Run() error {
	log.Println("Running path rename action")
	err := filepath.Walk(a.Dir,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			if err := a.skipExcluded(info.Name()); err != nil {
				return err
			}
			if err := a.processDir(path); err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}
