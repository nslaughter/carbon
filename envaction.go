package main

import (
	"errors"
	"fmt"
)

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
