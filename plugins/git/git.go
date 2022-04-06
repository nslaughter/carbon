package git

import (
	"errors"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"

	"github.com/nslaughter/carbon/framework"
)

func init() {
	framework.Register("git", New)
}

func New() framework.Action {
	return &GitAction{}
}

type GitAction struct {
	Source  string
	Dest    string
	Version string
	Command string
}

func (a *GitAction) Set(s framework.ActionSpec) error {
	return s.ToAction(a)
}

func (a *GitAction) Validate() error {
	return nil
}

func (a *GitAction) Run() error {
	switch a.Command {
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
		log.Printf("git command: %s not implemented", a.Command)
		os.Exit(1)
	}
	return nil
}
