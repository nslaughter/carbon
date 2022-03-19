package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func NewShellAction() Action {
	return &ShellAction{
		Commands: make([][]string, 0),
	}
}

type ShellAction struct {
	Dir      string
	Commands [][]string
}

func (a *ShellAction) Set(s ActionSpec) error {
	// set vars
	sm, err := ToStringsMap(s.Vars())
	if err != nil {
		return err
	}
	for k, v := range sm {
		switch k {
		case "dir":
			if d, ok := sm[k]; ok {
				a.Dir, err = Resolve(d)
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
		return fmt.Errorf("expected %T", ifcs)
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

func (a *ShellAction) Validate() error {
	return nil
}

func (a *ShellAction) Run() error {
	for _, c := range a.Commands {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = a.Dir
		out, err := cmd.Output()
		if err != nil {
			return errors.New("ShellAction: " + strings.Join(c, " "))
		}
		log.Println(out)
	}
	return nil
}
