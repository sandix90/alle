package models

import (
	"fmt"
)

type IManifestor interface {
	String() string
	GetFileName() string
	GetFullName() string
	GetTemplatePath() string
}

type Manifest struct {
	// name of the manifest. Usually is a file name.
	Name string `validate:"required"`

	// slice of manifest files paths which contain vars for manifest template
	VarsFilePaths []string `yaml:"vars"`

	// path to template
	templatePath string

	// deserialized template values which have been read from vars files. Package value could be mixed.
	values TemplateValues

	// templator interface which allows to parse values or create string manifest
	//templator services.templator

	// String manifest
	manifest string

	pack *Package
}

func NewManifest(name string, varFilePaths []string, templatePath string, manifestValues TemplateValues,
	pack *Package, stringManifest string) (*Manifest, error) {

	return &Manifest{
		Name:          name,
		VarsFilePaths: varFilePaths,
		values:        manifestValues,
		pack:          pack,
		templatePath:  templatePath,
		manifest:      stringManifest,
	}, nil
}

func (m *Manifest) String() string {
	return m.manifest
}

func (m *Manifest) GetFileName() string {
	return m.Name
}

func (m Manifest) GetFullName() string {
	return fmt.Sprintf("%s-%s", m.pack.Name, m.Name)
}

func (m Manifest) GetTemplatePath() string {
	return m.templatePath
}
