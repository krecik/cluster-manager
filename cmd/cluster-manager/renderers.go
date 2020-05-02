package main

import (
	"bytes"
	"fmt"
	"github.com/markbates/pkger"
	"io/ioutil"
	"os"
	"text/template"
)

func renderTemplate(path string, input interface{}) {
	file, err := pkger.Open(path)
	if err != nil {
		fatal(err)
	}

	templateBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fatal(err)
	}

	tmpl, err := template.New("inline").Parse(string(templateBytes))
	if err != nil {
		fatal(err)
	}

	err = tmpl.Execute(os.Stdout, input)
	if err != nil {
		fatal(err)
	}

	fmt.Println("---")
}

func renderTemplateToString(path string, input interface{}) string {
	var buffer bytes.Buffer

	file, err := pkger.Open(path)
	if err != nil {
		fatal(err)
	}

	templateBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fatal(err)
	}

	tmpl, err := template.New("inline").Parse(string(templateBytes))
	if err != nil {
		fatal(err)
	}

	err = tmpl.Execute(&buffer, input)
	if err != nil {
		fatal(err)
	}

	return buffer.String()
}
