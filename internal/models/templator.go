package models

import (
	"bytes"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

type Templator interface {
	ParseValues(values *TemplateValues, file string) error
	CreateTemplate(tmplFilePath string, buffer *bytes.Buffer, values *TemplateValues) error
}

type TemplatorImpl struct{}

func NewTemplator() *TemplatorImpl {
	return &TemplatorImpl{}
}

func (t *TemplatorImpl) ParseValues(values *TemplateValues, file string) error {
	filename, err := filepath.Abs(file)
	if err != nil {
		log.Error("Read file error")
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("Can't read file %s", filename)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &values.Values)
	if err != nil {
		return err
	}
	return nil
}

func (t *TemplatorImpl) CreateTemplate(tmplFilePath string, buffer *bytes.Buffer, values *TemplateValues) error {
	tmpl := template.Must(template.ParseFiles(tmplFilePath))
	err := tmpl.Execute(buffer, values)
	if err != nil {
		return err
	}
	return nil
}
