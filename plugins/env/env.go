package env

import (
	"github.com/nslaughter/carbon/framework"
)

type EnvAction struct{}

func init() {
	framework.Register("env", New)
}

func New() framework.Action {
	return &EnvAction{}
}

func (a *EnvAction) Set(s framework.ActionSpec) error {
	return framework.SetGlobalEnv(s)
}

func (a *EnvAction) Run() error {
	return nil
}

func (a *EnvAction) Validate() error {
	return nil
}
