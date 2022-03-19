package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/mitchellh/mapstructure"
)

var (
	defaultPath = "carbon.yaml"
	globalEnv   = make(map[string]string)
	registry    = make(map[string]ActionMaker)
)

// An ActionSpec contains data specifying an action.
type ActionSpec map[interface{}]interface{}

// Vars extracts the "vars" key from an ActionSpec.
func (s ActionSpec) Vars() interface{} {
	return s["vars"]
}

// ActionType names the action
func (s ActionSpec) ActionType() string {
	log.Println(s)
	for k := range s {
		return k.(string)
	}
	return ""
}

// Get either returns the value of a provided key and true or nil and false.
func (s ActionSpec) Get(k string) (interface{}, bool) {
	v, ok := s[k]
	return v, ok
}

func (s ActionSpec) ToAction(a interface{}) error {
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

// A Step is one unit of execution.
type Step map[string]interface{}

// Name should uniquely describe a Step.
func (s Step) Name() string {
	for k, v := range s {
		if k == "name" {
			return v.(string)
		}
	}
	return ""
}

func (s Step) ActionType() string {
	for k := range s {
		if k != "name" {
			return k
		}
	}
	return ""
}

func (s Step) ActionSpec() ActionSpec {
	return s[s.ActionType()].(map[interface{}]interface{})
}

// Script is the root of a carbon app
type Script struct {
	Name  string
	Env   map[string]string
	Steps []Step
}

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

// ============================================================================

// Actions contain the logic of script steps.
type Action interface {
	Set(s ActionSpec) error
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

// RegisterActions puts actions in the registry, so they're dicsoverable by executors
func RegisterActions() {
	registry["shell"] = NewShellAction
	registry["env"] = NewEnvAction
	registry["git"] = NewGitAction
	registry["text_replace"] = NewTextReplaceAction
	registry["path_rename"] = NewPathRenameAction
}

// Lookup wraps the env (pkg scope), so we don't contaminate too much code with package scoped var
func Lookup(k string) (string, error) {
	v, ok := globalEnv[k]
	if !ok {
		return "", errors.New(fmt.Sprintf("variable not present %s", k))
	}
	return v, nil
}

// Resolve names in env. Looks up when they have the $ prefix, else returns name.
func Resolve(v string) (string, error) {
	if strings.HasPrefix(v, "$") {
		return Lookup(v[1:])
	}
	return v, nil
}

// ============================================================================

// built-in keys like "env" and "workflow" could be plugins
// if we exposed access to environment in the framework API

func run() {
	var workflowName string
	var script Script

	RegisterActions()

	// TODO start taking actual input from CLI flags
	b, err := load(File)
	if err != nil {
		log.Println("failed loading script: ", err)
		os.Exit(1)
	}

	// get script
	if err := yaml.Unmarshal(b, &script); err != nil {
		log.Println("failed unmarshaling script: ", err)
		os.Exit(1)
	}

	log.Println("Executing script: ", script.Name)

	// process env vars
	for k, v := range script.Env {
		globalEnv[k] = v
	}

	for _, step := range script.Steps {
		log.Println("running: ", step.Name())
		log.Println("spec: ", step.ActionType())
		action, ok := registry[step.ActionType()]
		if !ok {
			log.Println("no plugin named ", step.ActionType())
		} else {
			a := action()
			if err := a.Set(step.ActionSpec()); err != nil {
				log.Println("setting: ", err)
				os.Exit(1)
			}
			if err := a.Run(); err != nil {
				log.Println("failed running: ", err)
				os.Exit(1)
			}
		}
	}

	log.Println("workflow: ", workflowName)
	log.Println("COMPLETE")
}
