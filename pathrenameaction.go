package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func NewPathRenameAction() Action {
	return &PathRenameAction{
		Excludes:      make([]string, 0),
		Substitutions: make([]Substitution, 0),
	}
}

// type Substitution struct {
// 	Old, New string
// }

type PathRenameAction struct {
	Dir           string
	Excludes      []string
	Substitutions []Substitution
}

func (a *PathRenameAction) Set(s ActionSpec) error {
	// set what we have at the top-level
	if err := mapstructure.Decode(s, a); err != nil {
		return fmt.Errorf("decoding top-level %w", err)
	}

	// get vars, resolve in global map, then decode them to top-level
	sm, ok := s.Vars().(map[interface{}]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("could not assert %T to %T", s.Vars(), sm))
	}

	for k, v := range sm {
		if sym, ok := v.(string); ok {
			sval, err := Resolve(sym)
			if err != nil {
				return err
			}
			sm[k] = sval
		}
	}

	if err := mapstructure.Decode(sm, a); err != nil {
		return fmt.Errorf("decoding vars %w", err)
	}

	return nil
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
