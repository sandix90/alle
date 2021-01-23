package services

import (
	"alle/internal"
	"alle/internal/models"
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Configurator interface {
	ParseConfig(aleConfig *models.AlleConfig, environment string, fileReader io.Reader) error
	findPackageByLabel(pack *models.Package, labels []string) bool
	GetStringManifestsByLabels(aleConfig *models.AlleConfig, labels []string) ([]string, error)
	GetStringPackageManifests(pack *models.Package) ([]string, error)
	GetPackagesByLabels(alleConfig *models.AlleConfig, labels []string) []*models.Package
}

type ConfiguratorImpl struct {
	templator Templator
}

func NewConfigurator(templator Templator) Configurator {
	return &ConfiguratorImpl{
		templator: templator,
	}
}

func (configurator *ConfiguratorImpl) ParseConfig(alleConfig *models.AlleConfig, environment string, configReader io.Reader) error {

	var b bytes.Buffer
	tmpl := template.New("config_template")

	fileStr, err := ioutil.ReadAll(configReader)
	if err != nil {
		return err
	}

	tmpl, err = tmpl.Parse(string(fileStr))
	if err != nil {
		return err
	}

	err = tmpl.Execute(&b, nil)
	if err != nil {
		return err
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

			if found := findPackageDuplicates(release.Packages, pack.Name); found {
				return fmt.Errorf("duplicate packages names not allowed. Duplicated pack name: \"%s\"", pack.Name)
			}

			packageValues, err := configurator.calculateTemplateValuesByFilePaths(pack.VarsFilePaths)
			if err != nil {
				return err
			}
			pack.SetPackageValues(packageValues)

			var newManifests []*models.Manifest
			for _, manifest := range pack.Manifests {

				manifestValues, err := configurator.calculateTemplateValuesByFilePaths(manifest.VarsFilePaths)
				if err != nil {
					return err
				}

				mergedValues := configurator.templator.MergeValues(packageValues, manifestValues)
				templatePath := fmt.Sprintf("%s/manifests/%s", pack.Path, manifest.Name)

				// Prepare string manifest
				templateFileReader, err := os.Open(templatePath)
				if err != nil {
					return err
				}

				var tmplWriter bytes.Buffer
				err = configurator.templator.RenderTemplate(templateFileReader, &tmplWriter, mergedValues)
				if err != nil {
					return err
				}

				err = templateFileReader.Close()
				if err != nil {
					return err
				}

				manifest, err = models.NewManifest(
					manifest.Name,
					manifest.VarsFilePaths,
					templatePath,
					mergedValues,
					pack,
					tmplWriter.String(),
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
		for _, packageVarFile := range pack.VarsFilePaths {
			packageValues := models.TemplateValues{}

			file, err := os.Open(packageVarFile)
			if err != nil {
				return nil, err
			}
			err = configurator.templator.ParseValues(&packageValues, bufio.NewReader(file))
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
		for _, manifestVarFile := range manifest.VarsFilePaths {
			manifestValues := models.TemplateValues{}

			file, err := os.Open(manifestVarFile)
			if err != nil {
				return nil, err
			}
			err = configurator.templator.ParseValues(&manifestValues, bufio.NewReader(file))
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
		//err = configurator.templator.RenderTemplate(file, &tmplBytes, &finalValues)
		//if err != nil {
		//	return nil, err
		//}
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

func (configurator *ConfiguratorImpl) calculateTemplateValuesByFilePaths(filePaths []string) (models.TemplateValues, error) {

	finalValues := models.TemplateValues{}

	for _, varsFilePath := range filePaths {
		file, err := os.Open(varsFilePath)
		if err != nil {
			return models.TemplateValues{}, err
		}

		// Read values from vars file
		lv := models.TemplateValues{}
		err = configurator.templator.ParseValues(&lv, file)
		if err != nil {
			return models.TemplateValues{}, err
		}

		finalValues = configurator.templator.MergeValues(finalValues, lv)

		err = file.Close()
		if err != nil {
			return models.TemplateValues{}, err
		}
	}

	return finalValues, nil
}

func findPackageDuplicates(packs []*models.Package, name string) bool {
	var foundCount uint8
	for _, p := range packs {
		if p.Name == name {
			foundCount++
		}

		if foundCount > 1 {
			return true
		}
	}
	return false
}
