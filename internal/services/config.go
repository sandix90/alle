package services

import (
	"alle/internal"
	"alle/internal/models"
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type Configurator interface {
	ParseConfig(aleConfig *models.AlleConfig, environment, filepath string) error
	findPackageByLabel(pack *models.Package, labels []string) bool
	GetStringManifestsByLabels(aleConfig *models.AlleConfig, labels []string) ([]string, error)
	GetStringPackageManifests(pack *models.Package) ([]string, error)
	GetPackagesByLabels(alleConfig *models.AlleConfig, labels []string) []*models.Package
}

type ConfiguratorImpl struct {
	Templator models.Templator
}

func NewConfigurator(templator models.Templator) *ConfiguratorImpl {
	return &ConfiguratorImpl{Templator: templator}
}

func (configurator *ConfiguratorImpl) ParseConfig(alleConfig *models.AlleConfig, environment, filepath string) error {

	workDir, err := os.Getwd()
	log.Debugf("Workdir: %s", workDir)
	log.Debugf("Using alle file: %s", filepath)
	err = internal.Exists(filepath)
	if err != nil {
		return fmt.Errorf("alle file is not found")
	}
	var b bytes.Buffer

	tmpl := template.Must(template.ParseFiles(filepath))
	if tmpl != nil {
		err := tmpl.Execute(&b, nil)
		if err != nil {
			return err
		}
	}
	err = yaml.Unmarshal(b.Bytes(), alleConfig)
	if err != nil {
		return err
	}
	alleConfig.Environment = environment

	err = internal.ValidateStruct(alleConfig)
	if err != nil {
		return err
	}

	for _, release := range alleConfig.Releases {
		for _, pack := range release.Packages {

			err = pack.SetPackageValues(configurator.Templator)
			if err != nil {
				return fmt.Errorf("error set package values. Package name: %s, path: %s", pack.Name, pack.Path)
			}
			var newManifests []*models.Manifest

			for _, manifest := range pack.Manifests {

				manifest, err = models.NewManifest(
					manifest.Name,
					manifest.Vars,
					configurator.Templator,
					pack.GetPackageValues(),
					pack.Path,
					pack,
				)
				newManifests = append(newManifests, manifest)

				if err != nil {
					return err
				}
			}
			pack.Manifests = newManifests
		}
	}

	return nil
}

func (configurator *ConfiguratorImpl) findPackageByLabel(pack *models.Package, labels []string) bool {
	for l, v := range pack.Labels {
		for _, al := range labels {
			parts := strings.Split(al, "=")
			if l == parts[0] && v == parts[1] {
				return true
			}
		}
	}
	return false
}

func (configurator *ConfiguratorImpl) GetStringManifestsByLabels(aleConfig *models.AlleConfig, labels []string) ([]string, error) {
	var tmpls []string
	for _, release := range aleConfig.Releases {
		//
		//out, err := release.GetStringPreConfigManifests()
		//log.Debugln(out)
		var err error
		for _, pack := range release.Packages {
			labelFound := configurator.findPackageByLabel(pack, labels)
			if labelFound {
				tmpls, err = configurator.GetStringPackageManifests(pack)
				if err != nil {
					return nil, err
				}
				fmt.Println(tmpls)
			}
		}
	}
	return tmpls, nil
}

func (configurator *ConfiguratorImpl) GetStringPackageManifests(pack *models.Package) ([]string, error) {
	var templatesOutput []string

	for _, manifest := range pack.Manifests {

		file := filepath.Join(pack.Path, "manifests", manifest.Name)
		err := internal.Exists(file)
		if err != nil {
			return nil, err
		}

		finalValues := models.TemplateValues{}
		// Get package values
		for _, packageVarFile := range pack.Vars {
			packageValues := models.TemplateValues{}

			err := configurator.Templator.ParseValues(&packageValues, packageVarFile)
			if err != nil {
				return nil, err
			}
			if finalValues.Values != nil {
				finalValues.Values = internal.MergeMaps(finalValues.Values, packageValues.Values)
			} else {
				finalValues = packageValues
			}
		}

		// Get manifest values
		for _, manifestVarFile := range manifest.Vars {
			manifestValues := models.TemplateValues{}

			err := configurator.Templator.ParseValues(&manifestValues, manifestVarFile)
			if err != nil {
				return nil, err
			}
			if finalValues.Values != nil {
				finalValues.Values = internal.MergeMaps(finalValues.Values, manifestValues.Values)
			} else {
				finalValues = manifestValues
			}
		}

		var tmplBytes bytes.Buffer
		err = configurator.Templator.CreateTemplate(file, &tmplBytes, &finalValues)
		if err != nil {
			return nil, err
		}
		templatesOutput = append(templatesOutput, tmplBytes.String())
	}
	return templatesOutput, nil
}

func (configurator *ConfiguratorImpl) GetPackagesByLabels(alleConfig *models.AlleConfig, labels []string) []*models.Package {
	var foundPackages []*models.Package

	for _, release := range alleConfig.Releases {
		for _, pack := range release.Packages {
			if configurator.findPackageByLabel(pack, labels) {
				foundPackages = append(foundPackages, pack)
			}
		}
	}
	return foundPackages
}
