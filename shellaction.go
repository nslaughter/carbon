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

func (a *ShellAction) Set(vars interface{}) error {
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
			if d, ok := varMap["dir"]; ok {
				a.Dir, err = Resolve(d)
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
					return errors.New("not quite")
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

func (a *ShellAction) Validate() error {
	return nil
}

func (a *ShellAction) Run() error {
	for _, c := range a.Commands {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = a.Dir
		out, err := cmd.Output()
		log.Println(a.Dir, out)
		if err != nil {
			log.Println("command: ", c, " - msg: ", err)
		}
		log.Println(out)
	}
	return nil
}
