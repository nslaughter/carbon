package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TODO this is almost identical to the text_replace action. Their respective
// processXXX methods are the differences.

func NewPathRenameAction() Action {
	return &PathRenameAction{
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
	// set vars
	sm, ok := s.Vars().(map[interface{}]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("could not assert %T to %T", s.Vars(), sm))
	}
	for k, v := range sm {
		switch k {
		case "dir":
			if d, ok := sm[k]; ok {
				if sk, ok := d.(string); ok {
					var err error
					if a.Dir, err = Resolve(sk); err != nil {
						return err
					}
				}
			}
		case "exclude":
			if d, ok := sm[k]; ok {
				ifcs, ok := d.([]interface{})
				if !ok {
					return errors.New("could not assert exclude interface{}")
				}
				for _, v := range ifcs {
					s, ok := v.(string)
					if !ok {
						return errors.New("could not assert " + string(s))
					}
					a.Excludes = append(a.Excludes, s)
				}
			}
		default:
			log.Println("unexpected var key: ", k, "; value: ", v)
		}
	}

	i, ok := s.Get("substitutions")
	if !ok {
		return errors.New("key not found: substitutions")
	}
	ifcs, ok := i.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("could not assert %T to %T", i, ifcs))
	}
	for _, v := range ifcs {
		oldIfc, _ := getKey(v, "old")
		newIfc, _ := getKey(v, "new")
		oldStr, oldOK := oldIfc.(string)
		newStr, newOK := newIfc.(string)
		if !oldOK || !newOK {
			return errors.New(fmt.Sprintf("could not assert %T | %T to string | string - %s", old, new, oldStr))
		}
		s := Substitution{Old: oldStr, New: newStr}
		a.Substitutions = append(a.Substitutions, s)
		log.Println("Adding PATH substitution: ", s)
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
