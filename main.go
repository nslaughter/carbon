package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-yaml/yaml"
)

var (
	defaultPath = "carbon.yaml"
	globalEnv   = make(map[string]string)
	registry    = make(map[string]ActionMaker)
)

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

// RegisterActions puts actions in the registry, so they're dicsoverable by executors
func RegisterActions() {
	registry["shell"] = NewShellAction
	registry["env"] = NewEnvAction
	registry["git"] = NewGitAction
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

func main() {
	var workflowName string
	script := make(map[string]interface{})
	// env := make(map[string]string)

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
