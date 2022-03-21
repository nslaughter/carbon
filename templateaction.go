package main

import (
    "bytes"
    "log"
	"text/template"
    "time"
)

func NewTemplateAction() Action {
	return &TemplateAction{}
}

type TemplateAction struct {
	Content string
	Source  string
	Dest    string
	Data    map[interface{}]interface{}
}

func (a *TemplateAction) Set(s ActionSpec) error {
	return s.ToAction(a)
}

func (a *TemplateAction) Validate() error {
	return nil
}

func (a *TemplateAction) Run() error {
    buf := bytes.NewBuffer([]byte(""))
	tmpl, err := template.New("demo").Parse("{{.first}} {{.second}} {{.third}}")
	if err != nil {
		return err
	}
	err = tmpl.Execute(buf, a.Data)
	if err != nil {
		return err
	}
    log.Println(buf.String())
    time.Sleep(100 * time.Millisecond)
	return nil
}
