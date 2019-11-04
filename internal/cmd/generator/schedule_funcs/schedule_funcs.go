// The following directive is necessary to make the package coherent:
// This program generates types, It can be invoked by running
// go generate
package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
)

type task struct {
	Name        string   `json:"-"`
	Description string   `json:"description"`
	Value       []string `json:"value,omitempty"`
}

var funcs = template.FuncMap{
	"lowerFirst": func(s string) string {
		if len(s) == 0 {
			return ""
		}
		if s[0] < 'A' || s[0] > 'Z' {
			return s
		}
		return string(s[0]+'a'-'A') + s[1:]
	},
	"endwith": func(x, y string) bool {
		return strings.HasSuffix(x, y)
	},
	"merge": func(x, y []string) []string {
		a := make(map[string]struct{})
		for _, v := range x {
			a[v] = struct{}{}
		}
		for _, v := range y {
			a[v] = struct{}{}
		}
		o := make([]string, 0)
		for x := range a {
			o = append(o, x)
		}

		sort.Strings(o)
		return o
	},
}

func executeTemplate(tmpl *template.Template, w io.Writer, v interface{}) {
	err := tmpl.Execute(w, v)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	data, err := ioutil.ReadFile("schedule_func.json")
	if err != nil {
		log.Fatal(err)
	}

	tasks := make(map[string]*task)
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	// Do sort to all tasks via name.
	taskNames := make([]string, 0)
	for k := range tasks {
		sort.Strings(tasks[k].Value)

		taskNames = append(taskNames, k)
	}
	sort.Strings(taskNames)

	// Format input tasks.json
	data, err = json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("schedule_func.json", data, 0664)
	if err != nil {
		log.Fatal(err)
	}

	// Set task names.
	for k, v := range tasks {
		v.Name = k
	}

	taskFile, err := os.Create("schedule_func.go")
	if err != nil {
		log.Fatal(err)
	}
	defer taskFile.Close()

	executeTemplate(requirementPageTmpl, taskFile, nil)

	for _, taskName := range taskNames {
		v := tasks[taskName]

		executeTemplate(requirementTmpl, taskFile, v)
	}
}

var requirementPageTmpl = template.Must(template.New("requirementPage").Funcs(funcs).Parse(`// Code generated by go generate; DO NOT EDIT.
package types

import (
	"github.com/Xuanwo/navvy"
)
`))

var requirementTmpl = template.Must(template.New("requirement").Funcs(funcs).Parse(`
// {{ .Name }}Requirement is the requirement for {{ .Name }}Task.
type {{ .Name }}Requirement interface {
	navvy.Task

	// Value
{{- range $k, $v := .Value }}
	{{$v}}Getter
	{{$v}}Setter
	{{$v}}Validator
{{- end }}
}

type {{ .Name | lowerFirst }}ScheduleFunc func(navvy.Task){{ .Name }}Requirement
`))