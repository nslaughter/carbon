package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-yaml/yaml"
)

var (
	defaultPath = "carbon.yaml"
	globalEnv   = make(map[string]string)
	registry    = make(map[string]ActionMaker)
)

// ============================================================================

// TODO finish implementing all "built-ins" as ordinary commands
// TODO organize config for execution (Options pattern?)
// TODO migrate code to Github repo
// TODO setup Github project
// TODO implement env as framework API exposed for plugs
// TODO make tool level config (can override in script)
// TODO standardize variable parsing/setting for Actions in lib
// TODO design change for unmarshalling actions with order preserved
// TODO design template builder step
// TODO write action to validate script requirements are on host
// TODO build initial CLI
// TODO add CICD / Github Action

// ============================================================================

// load slurps a file. If path is empty string it looks for carbon.yaml
func load(path string) ([]byte, error) {
	if path == "" {
		path = defaultPath
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type Value interface{}

type Key string

// ============================================================================

type Action interface {
	Set(vars interface{}) error
	Validate() error
	Run() error
}

type ActionMaker func() Action

// ToStringsMap is a utility function for converting interface{} to map[string]string
func ToStringsMap(in interface{}) (map[string]string, error) {
	res := make(map[string]string)
	inMap, ok := in.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("key not string")
	}
	for k, v := range inMap {
		kVal, kOK := k.(string)
		vVal, vOK := v.(string)
		if !kOK || !vOK {
			return res, errors.New(fmt.Sprintf("%T could not map to %T", in, res))
		}
		res[kVal] = vVal
	}
	return res, nil
}

// ============================================================================

// Lookup wraps the env (pkg scope), so we don't contaminate too much code with package scoped var
func Lookup(k string) (string, error) {
	v, ok := globalEnv[k]
	if !ok {
		return "", errors.New(fmt.Sprintf("variable not present %s", k))
	}
	return v, nil
}

// Resolve names in env when thye have the $ prefix
func Resolve(v string) (string, error) {
	if strings.HasPrefix(v, "$") {
		return Lookup(v[1:])
	}
	return v, nil
}

// ============================================================================

type EnvAction struct{}

func NewEnvAction() Action {
	return &EnvAction{}
}

func (a *EnvAction) Set(vars interface{}) error {
	attrs, ok := vars.(map[interface{}]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("vars %T not %T", vars, attrs))
	}
	for k, v := range attrs {
		kstr, kOK := k.(string)
		vstr, vOK := v.(string)
		if !kOK || !vOK {
			return errors.New("key not string")
		}
		globalEnv[kstr] = vstr
	}
	return nil
}

func (a *EnvAction) Run() error {
	return nil
}

func (a *EnvAction) Validate() error {
	return nil
}

// ============================================================================

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

// ============================================================================

// built-in keys like "env" and "workflow" could be plugins
// if we exposed access to environment in the framework API

func main() {
	var workflowName string
	script := make(map[string]interface{})
	// env := make(map[string]string)

	// crude registration mechanism
	registry["shell"] = NewShellAction
	registry["env"] = NewEnvAction

	// TODO start taking actual input from CLI flags
	b, err := load("")
	if err != nil {
		log.Println("failed loading script: ", err)
		os.Exit(1)
	}

	// get script
	if err := yaml.Unmarshal(b, &script); err != nil {
		log.Println("failed unmarshaling script: ", err)
		os.Exit(1)
	}

	// first check built ins, default goes to the plugin registry
	for k, v := range script {
		switch k {
		case "workflow":
			name, ok := v.(string)
			if !ok {
				log.Printf("expected workflow to be %T", name)
			}
			workflowName = name
			continue
		default:
			log.Println("doing action: ", k)
			action, ok := registry[k]
			if !ok {
				log.Println("no plugin named ", k)
			} else {
				venv, ok := v.(map[interface{}]interface{})
				if !ok {
					log.Printf("action block NO")
					os.Exit(1)
				}
				a := action()
				if err := a.Set(venv); err != nil {
					log.Println("setting: ", err)
					os.Exit(1)
				}
				if err := a.Run(); err != nil {
					log.Println("failed running: ", err)
					os.Exit(1)
				}
			}
		}
		log.Printf(k)
	}

	log.Println("workflow: ", workflowName)
	log.Println("COMPLETE")
}
