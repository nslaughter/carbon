package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

// Template scripts are the most complicated feature.
func NewTemplateAction() Action {
	return &TemplateAction{
		Force:    true,
		Backup:   false,
		Excludes: make([]string, 0),
		Includes: make([]string, 0),
	}
}

// TemplateAction expands template provided source, dest, and data.
type TemplateAction struct {
	// Force overwrites Dest if it exists and if Force then Backup will
	// create a backup to restore in case you incorrectly overwrite.
	Force, Backup bool
	// Source is checked first and if it's not empty then Content is ignored.
	// Content can contain a string value for applying templates instead of paths.
	Source, Content string
	// Dest is the path for template output.
	Dest string
	// Data contains values from external file.
	Data map[interface{}]interface{}

	Excludes []string
	Includes []string
}

func (a *TemplateAction) Set(s ActionSpec) error {
	return s.ToAction(a)
}

func (a *TemplateAction) Validate() error {
	return nil
}

// OPTIMIZATION we could precompile regexps instead of repeating the effort for each path
func multimatch(patterns []string, s string) (matched bool, err error) {
	for _, p := range patterns {
		matched, err = regexp.MatchString(p, s)
		if err != nil || matched {
			return
		}
	}
	return
}

func parseFile(tmpl *template.Template, fullpath, basepath string) (*template.Template, error) {
	b, err := os.ReadFile(fullpath)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %w", err)
	}

	rel, err := filepath.Rel(basepath, fullpath)
	if err != nil {
		return nil, fmt.Errorf("Rel: %w", err)
	}

	tmpl, err = tmpl.New(rel).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}

	return tmpl, nil
}

func (a *TemplateAction) filterDir(path string) error {
	matched, err := multimatch(a.Excludes, path)
	if err != nil {
		return fmt.Errorf("multimatch: %w", err)
	}
	if matched {
		return filepath.SkipDir
	}
	return nil
}

// parseDir builds templates walking the directory and names templates for their
// path. This keeps full path association with each template node.
func (a *TemplateAction) parseDir(path string) (*template.Template, error) {
	tmpl := template.New(path)
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			log.Println("walk path: ", path)
			if err != nil {
				return fmt.Errorf("walk fail: %w", err)
			}
			if info.IsDir() {
				return a.filterDir(path)
			}

			matched, err := multimatch(a.Excludes, path)
			if err != nil {
				return fmt.Errorf("multimatch: %w", err)
			}
			if matched {
				return nil
			}

			tmpl, err = parseFile(tmpl, path, a.Source)
			if err != nil {
				return err
			}

			return nil
		})
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

/*
func (a *TemplateAction) parse() (*template.Template, error) {
	// when source is present we use it
	var (
		tmplname string
		tmplstr  string
	)

	if a.Source != "" {
		info, err := os.Stat(a.Source)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return a.parseDir(info.Name())
		}

		b, err := os.ReadFile(a.Source)
		if err != nil {
			return nil, err
		}
		tmplstr = string(b)
		tmplname = a.Source
	} else {
		tmplstr = a.Content
		tmplname = "content"
	}

	if tmplstr == "" {
		return nil, errors.New("no template provided")
	}
	t, err := template.New(tmplname).Parse(tmplstr)
	if err != nil {
		return nil, err
	}
	return t, nil
}
*/

// executeContent executes content as a template, with data, to the dest
func executeContent(dest, content string, data interface{}) error {
	dir := filepath.Dir(dest)
	if _, err := os.Stat(dir); err != nil {
		if err != os.ErrNotExist {
			return err
		}
		if err := os.Mkdir(dir, 0o777); err != nil {
			return nil
		}
		return err
	}

	tmpl, err := template.New("content").Parse(content)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0o664)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return err
	}

	return nil
}

// executeTmpl creates the directory needed
func (a *TemplateAction) executeTmpl(t *template.Template, data interface{}) error {
	dir := filepath.Dir(t.Name())
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(filepath.Join(a.Dest, t.Name()), os.O_RDWR|os.O_CREATE, 0o664)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		return err
	}

	return nil
}

func (a *TemplateAction) Run() error {
	// 3 valid cases of Source field:
	// (1) content provided, (2) path to file, (3) path to dir
	if a.Content != "" {
		return executeContent(a.Dest, a.Content, a.Data)
	}

	info, err := os.Stat(a.Source)
	if err != nil {
		return fmt.Errorf("could not stat Source: %w", err)
	}

	if info.IsDir() {
		t, err := a.parseDir(a.Source)
		if err != nil {
			return fmt.Errorf("parseDir: %w", err)
		}
		for _, v := range t.Templates() {
			// TODO execute templates into path
			println("executing: ", v.Name())
			if err := a.executeTmpl(v, a.Data); err != nil {
				return fmt.Errorf("could not execute template: %w", err)
			}
		}
		return nil
	}
	return nil
}
