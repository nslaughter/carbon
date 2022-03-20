package main

import (
	"errors"
	"fmt"
)

type EnvAction struct{}

func NewEnvAction() Action {
	return &EnvAction{}
}

func (a *EnvAction) Set(s ActionSpec) error {
	attrs, ok := interface{}(s).(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("vars %T not %T", s.Vars(), attrs)
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
