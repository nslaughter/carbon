package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func NewTextReplaceAction() Action {
	return &TextReplaceAction{
		Excludes:      make([]string, 0),
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
