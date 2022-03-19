package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
)

func NewGitAction() Action {
	return &GitAction{}
}

type GitAction struct {
	Source   string
	Dest     string
	Version  string
	Commands [][]string
}

func (a *GitAction) Set(s ActionSpec) error {
	// set vars
	sm, err := ToStringsMap(s.Vars())
	if err != nil {
		return err
	}
	for k, v := range sm {
		switch k {
		case "source":
			if d, ok := sm[k]; ok {
				a.Source, err = Resolve(d)
			}
		case "dest":
			if d, ok := sm[k]; ok {
				a.Dest, err = Resolve(d)
			}
		default:
			log.Println("cannot handle key: ", k, "; value: ", v)
		}
	}

	// set commands
	incs, ok := s.Get("commands")
	if !ok {
		return errors.New("commands key not found")
	}
	cs := make([][]string, 0)
	ifcs, ok := incs.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("expected %T", ifcs))
	}
	for _, item := range ifcs {
		c, ok := item.(string)
		if !ok {
			return errors.New("expected string argument")
		}
		cs = append(cs, strings.Split(c, " "))
	}
	a.Commands = cs

	return nil
}

func (a *GitAction) Validate() error {
	return nil
}

func (a *GitAction) Run() error {
	for _, c := range a.Commands {
		switch c[0] {
		case "clone":
			if _, err := os.Stat(a.Dest); !errors.Is(err, os.ErrNotExist) {
				// TODO determine when this is fatal
				log.Println("destination path already exists")
				return nil
			}
			_, err := git.PlainClone(a.Dest, false, &git.CloneOptions{
				URL: a.Source,
			})
			if err != nil {
				return err
			}
		default:
			log.Printf("git command: %s not implemented", c[0])
			os.Exit(1)
		}
	}
	return nil
}
