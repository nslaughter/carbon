package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func NewTextReplaceAction() Action {
	return &TextReplaceAction{
		Substitutions: make([]Substitution, 0),
	}
}

type Substitution struct {
	Old, New string
}

type TextReplaceAction struct {
	Dir           string
	Excludes      []string
	Substitutions []Substitution
}

func (a *TextReplaceAction) Set(s ActionSpec) error {
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
			return errors.New(fmt.Sprintf("could not assert %T | %T to string | string - %s", oldIfc, newIfc, oldStr))
		}
		s := Substitution{Old: oldStr, New: newStr}
		a.Substitutions = append(a.Substitutions, s)
		log.Println("Adding TEXT substitution: ", s)
	}

	return nil
}

func getKey(m interface{}, k string) (interface{}, error) {
	v, ok := m.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("could not assert")
	}
	i, ok := v[k]
	if !ok {
		return nil, errors.New("no key " + k)
	}
	return i, nil
}

func (a *TextReplaceAction) Validate() error {
	return nil
}

func (a *TextReplaceAction) processFile(path string) error {
	b, err := os.ReadFile(path)
	for _, s := range a.Substitutions {
		if err != nil {
			return err
		}
		b = bytes.ReplaceAll(b, []byte(s.Old), []byte(s.New))
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	return nil
}

func (a *TextReplaceAction) skipExcluded(path string) error {
	for _, excl := range a.Excludes {
		if strings.Contains(path, excl) {
			return filepath.SkipDir
		}
	}
	return nil
}

func (a *TextReplaceAction) Run() error {
	err := filepath.Walk(a.Dir,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return a.skipExcluded(path)
			}
			if err := a.processFile(path); err != nil {
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
