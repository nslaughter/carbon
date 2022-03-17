package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
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

func (a *GitAction) Set(vars interface{}) error {
	attrs, ok := vars.(map[interface{}]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("vars %T not %T", vars, attrs))
	}
	for k, v := range attrs {
		kstr, ok := k.(string)
		if !ok {
			return errors.New("key not string")
		}

		switch kstr {
		case "after":
			// just here to catch dependency right now

		case "vars":
			varMap, err := ToStringsMap(v)
			if err != nil {
				return err
			}
			if d, ok := varMap["source"]; ok {
				a.Source, err = Resolve(d)
				if err != nil {
					return err
				}
			}
			if d, ok := varMap["dest"]; ok {
				a.Dest, err = Resolve(d)
				if err != nil {
					return err
				}
			}
			if d, ok := varMap["version"]; ok {
				a.Version, err = Resolve(d)
				if err != nil {
					return err
				}
			}
		case "commands":
			cs := make([][]string, 0)
			ifcs, ok := v.([]interface{})
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
		default:

			return errors.New("unexpected ")
		}
	}
	return nil
}

func (a *GitAction) Validate() error {
	return nil
}

func (a *GitAction) Run() error {
	for _, c := range a.Commands {
		switch c[0] {
		case "clone":
			// either use what's in the command expression or look for vars
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
