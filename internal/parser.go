package internal

import (
	"bytes"
	"github.com/ghodss/yaml"
	"html/template"
	"path/filepath"
	"strings"
)


type Wait struct {
	For string `validate:"required"`
	Condition string `validate:"required"`
	Timeout string
}

type Manifest struct {
	Name string `validate:"required"`
	Vars []string
}

type Schema struct {
	Path string `validate:"required"`
	Manifests []Manifest `validate:"dive"`
}

type Package struct {
	Name string `validate:"required"`
	Schema *Schema `validate:"dive"`
	Labels map[string]string
	Vars []string
	Wait *Wait
}

func (pack *Package) GetStringSchemaManifests() ([]string, error){
	var templatesOutput []string

	for _, manifest := range pack.Schema.Manifests {

		file := filepath.Join(pack.Schema.Path, "manifests", manifest.Name)
		err := Exists(file)
		if err != nil {
			return nil, err
		}

		finalValues := TemplateValues{}
		for _, varFile := range pack.Vars {
			localValues := TemplateValues{}

			err := ParseAlleValues(&localValues, varFile)
			if err != nil {
				return nil, err
			}
			if finalValues.Values != nil {
				*finalValues.Values = mergeMaps(*finalValues.Values, *localValues.Values)
			} else {
				finalValues = localValues
			}
		}

		for _, manifestVarFile := range manifest.Vars{
			localValues := TemplateValues{}

			err := ParseAlleValues(&localValues, manifestVarFile)
			if err != nil {
				return nil, err
			}
			if finalValues.Values != nil {
				*finalValues.Values = mergeMaps(*finalValues.Values, *localValues.Values)
			} else {
				finalValues = localValues
			}
		}

		var tmplBytes bytes.Buffer
		err = CreateTemplate(file, &tmplBytes, &finalValues)
		if err != nil {
			return nil, err
		}
		templatesOutput = append(templatesOutput, tmplBytes.String())
	}
	return templatesOutput, nil
}

type Release struct{
	Name string `json:"name" validate:"required"`
	Packages []*Package `validate:"dive"`
	PreConfig []*PreConfig `yaml:"pre_config" json:"pre_config" validate:"dive"`
}

func (release *Release) GetStringPreConfigManifests() ([]string, error){
	var out []string

	for _, preconfig := range release.PreConfig{
		if mans, err := preconfig.GetStringManifests(); err != nil {
			return nil, err
		} else {
			for _, man :=range mans{
				out = append(out, man)
			}

		}
	}
	return out, nil
}

type PreConfig struct {
	Name string `validate:"required"`
	Schema *Schema `validate:"dive"`
	Secrets string
	Order int
}

func (pc *PreConfig) GetStringManifests() ([]string, error){

	var templatesOutput []string

	for _, manifest := range pc.Schema.Manifests {

		file := filepath.Join(pc.Schema.Path, "manifests", manifest.Name)
		err := Exists(file)
		if err != nil {
			return nil, err
		}

		finalValues := TemplateValues{}
		for _, varFile := range manifest.Vars {
			localValues := TemplateValues{}

			err := ParseAlleValues(&localValues, varFile)
			if err != nil {
				return nil, err
			}
			if finalValues.Values != nil {
				*finalValues.Values = mergeMaps(*finalValues.Values, *localValues.Values)
			} else {
				finalValues = localValues
			}
		}

		for _, manifestVarFile := range manifest.Vars{
			localValues := TemplateValues{}

			err := ParseAlleValues(&localValues, manifestVarFile)
			if err != nil {
				return nil, err
			}
			if finalValues.Values != nil {
				*finalValues.Values = mergeMaps(*finalValues.Values, *localValues.Values)
			} else {
				finalValues = localValues
			}
		}

		var tmplBytes bytes.Buffer
		err = CreateTemplate(file, &tmplBytes, &finalValues)
		if err != nil {
			return nil, err
		}
		templatesOutput = append(templatesOutput, tmplBytes.String())
	}
	return templatesOutput, nil

}

type AlleConfig struct {
	Environment string `validate:"required"`
	Releases []*Release `yaml:"releases" validate:"dive"`
	GlobalPreConfig []*PreConfig `yaml:"pre_config" validate:"dive"`
}


func UnmarshalAlleConfig(aleConfig *AlleConfig, file string) error{
	var b bytes.Buffer

	tmpl := template.Must(template.ParseFiles(file))
	if tmpl != nil {
		err := tmpl.Execute(&b, nil)
		if err != nil {
			return err
		}
	}
	//aleConfig := AlleConfig{}
	err := yaml.Unmarshal(b.Bytes(), aleConfig)
	if err != nil {
		return err
	}

	err = ValidateStruct(aleConfig)
	if err != nil{
		return  err
	}
	return nil
}

func FindByLabel(pack *Package, labels []string) bool{
	for l, v := range pack.Labels{
		for _, al := range labels{
			parts := strings.Split(al, "=")
			if l == parts[0] && v == parts[1]{
				return true
			}
		}
	}
	return false
}
