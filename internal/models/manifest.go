package models

import (
	"alle/internal"
	"bytes"
	"fmt"
)

type IManifestor interface {
	String() (string, error)
	GetName() string
	GetFullPath() string
}

type Manifest struct {
	// name of the manifest. Usually is a file name.
	Name string `validate:"required"`

	// slice of manifest files paths which contain vars for manifest template
	Vars []string

	// path to template
	templatePath string

	// deserialized template values which have been read from vars files. Package value could be mixed.
	values *TemplateValues

	// Templator interface which allows to parse values or create string manifest
	templator Templator

	pack *Package
}

func NewManifest(name string, vars []string, templator Templator, packageValues TemplateValues, packagePath string, pack *Package) (*Manifest, error) {
	manifestTemplatePath := fmt.Sprintf("%s/manifests/%s", packagePath, name)

	// Parse manifest files with vars
	manifestValues := new(TemplateValues)

	// Merge with package values
	manifestValues.Values = packageValues.Values

	for _, varsFilePath := range vars {
		err := internal.Exists(varsFilePath)
		if err != nil {
			return nil, fmt.Errorf("vars file path doesn't exist. path: %s", varsFilePath)
		}

		varsFileValues := new(TemplateValues)
		err = templator.ParseValues(varsFileValues, varsFilePath)
		if err != nil {
			return nil, fmt.Errorf("error parse manifest values.\n manifest path: %s\nOrigin error: %w", manifestTemplatePath, err)
		}

		if manifestValues.Values != nil {
			manifestValues.Values = internal.MergeMaps(manifestValues.Values, varsFileValues.Values)
		} else {
			manifestValues = varsFileValues
		}
	}

	return &Manifest{
		Name:         name,
		Vars:         vars,
		templatePath: manifestTemplatePath,
		values:       manifestValues,
		templator:    templator,
		pack:         pack,
	}, nil
}

func (m *Manifest) String() (string, error) {

	var tmplBytes bytes.Buffer
	err := m.templator.CreateTemplate(m.templatePath, &tmplBytes, m.values)
	if err != nil {
		return "", err
	}

	return tmplBytes.String(), nil
}

func (m *Manifest) GetName() string {
	return m.Name
}

func (m Manifest) GetFullPath() string {
	return fmt.Sprintf("%s-%s", m.pack.Name, m.Name)
}
