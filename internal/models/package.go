package models

type Package struct {
	Name          string      `validate:"required"`
	Path          string      `validate:"required"`
	Manifests     []*Manifest `validate:"dive"`
	Labels        map[string]string
	VarsFilePaths []string `yaml:"vars"`
	Wait          *Wait

	packageValues TemplateValues
}

func (p *Package) SetPackageValues(values TemplateValues) {
	p.packageValues = values
}

func (p *Package) GetPackageValues() TemplateValues {
	return p.packageValues
}
