package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
	return s.ToAction(a)
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
	if err := os.WriteFile(path, b, 0o600); err != nil {
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
