package models

import (
	"alle/internal"
	"fmt"
)

type Package struct {
	Name      string      `validate:"required"`
	Path      string      `validate:"required"`
	Manifests []*Manifest `validate:"dive"`
	Labels    map[string]string
	Vars      []string
	Wait      *Wait

	packageValues TemplateValues
}

func (p *Package) SetPackageValues(templator Templator) error {
	// Parse package files with vars
	for _, varsFilePath := range p.Vars {
		err := internal.Exists(varsFilePath)
		if err != nil {
			return fmt.Errorf("vars file path doesn't exist. path: %s", varsFilePath)
		}

		varsFileValues := new(TemplateValues)
		err = templator.ParseValues(varsFileValues, varsFilePath)
		if err != nil {
			return fmt.Errorf("error parse manifest values.\n manifest path: %s\nOrigin error: %w", varsFileValues, err)
		}

		if p.packageValues.Values != nil {
			p.packageValues.Values = internal.MergeMaps(p.packageValues.Values, varsFileValues.Values)
		} else {
			p.packageValues = *varsFileValues
		}
	}
	return nil
}

func (p *Package) GetPackageValues() TemplateValues {
	return p.packageValues
}
