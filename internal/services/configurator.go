package services

import (
	"alle/internal"
	"alle/internal/models"
	"bufio"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Configurator interface {
	findPackageByLabel(pack *models.Package, labels []string) bool
	GetStringManifestsByLabels(labels []string) ([]string, error)
	GetStringPackageManifests(pack *models.Package) ([]string, error)
	GetPackagesByLabels(labels []string) []*models.Package
	GetAlleConfig() (*models.AlleConfig, error)
}

type configuratorImpl struct {
	templator  Templator
	alleConfig *models.AlleConfig
}

func NewConfiguratorFromFile(templator Templator, environment string, filepath string) (Configurator, error) {

	workDir, err := os.Getwd()
	log.Debugf("Workdir: %s", workDir)
	log.Debugf("Using alle file: %s", filepath)

	err = internal.Exists(filepath)
	if err != nil {
		return nil, fmt.Errorf("alle file is not found")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error open file: %s. OErr: %v", filepath, err)
	}

	configurator, err := NewConfigurator(templator, environment, file)
	if err != nil {
		return nil, fmt.Errorf("error creating configurator. OError: %w", err)
	}

	return configurator, nil
}

func NewConfigurator(templator Templator, environment string, configReader io.Reader) (Configurator, error) {
	alleConfig, err := parseConfig(templator, environment, configReader)
	if err != nil {
		return nil, fmt.Errorf("error parsing config. OError: %v", err)
	}
	return &configuratorImpl{
		templator:  templator,
		alleConfig: alleConfig,
	}, nil
}

func parseConfig(templator Templator, environment string, configReader io.Reader) (*models.AlleConfig, error) {

	var b bytes.Buffer
	alleConfig := new(models.AlleConfig)

	tmpl := template.New("config_template")

	fileStr, err := ioutil.ReadAll(configReader)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.Parse(string(fileStr))
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(&b, nil)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b.Bytes(), alleConfig)
	if err != nil {
		return nil, err
	}
	alleConfig.Environment = environment

	err = internal.ValidateStruct(alleConfig)
	if err != nil {
		return nil, err
	}

	for _, release := range alleConfig.Releases {
		for _, pack := range release.Packages {

			if found := findPackageDuplicates(release.Packages, pack.Name); found {
				return nil, fmt.Errorf("duplicate packages names not allowed. Duplicated pack name: \"%s\"", pack.Name)
			}

			packageValues, err := calculateTemplateValuesByFilePaths(templator, pack.VarsFilePaths)
			if err != nil {
				return nil, err
			}
			pack.SetPackageValues(packageValues)

			var newManifests []*models.Manifest
			for _, manifest := range pack.Manifests {

				manifestValues, err := calculateTemplateValuesByFilePaths(templator, manifest.VarsFilePaths)
				if err != nil {
					return nil, err
				}

				mergedValues := templator.MergeValues(packageValues, manifestValues)
				templatePath := fmt.Sprintf("%s/manifests/%s", pack.Path, manifest.Name)

				// Prepare string manifest
				templateFileReader, err := os.Open(templatePath)
				if err != nil {
					return nil, err
				}

				var tmplWriter bytes.Buffer
				err = templator.RenderTemplate(templateFileReader, &tmplWriter, mergedValues)
				if err != nil {
					return nil, err
				}

				err = templateFileReader.Close()
				if err != nil {
					return nil, err
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
					return nil, err
				}
			}
			pack.Manifests = newManifests
		}
	}

	return alleConfig, nil
}

func (configurator *configuratorImpl) GetAlleConfig() (*models.AlleConfig, error) {
	return configurator.alleConfig, nil
}

func (configurator *configuratorImpl) findPackageByLabel(pack *models.Package, labels []string) bool {
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

func (configurator *configuratorImpl) GetStringManifestsByLabels(labels []string) ([]string, error) {
	var tmpls []string
	for _, release := range configurator.alleConfig.Releases {
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

func (configurator *configuratorImpl) GetStringPackageManifests(pack *models.Package) ([]string, error) {
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

func (configurator *configuratorImpl) GetPackagesByLabels(labels []string) []*models.Package {
	var foundPackages []*models.Package

	for _, release := range configurator.alleConfig.Releases {
		for _, pack := range release.Packages {
			if configurator.findPackageByLabel(pack, labels) {
				foundPackages = append(foundPackages, pack)
			}
		}
	}
	return foundPackages
}

func calculateTemplateValuesByFilePaths(templator Templator, filePaths []string) (models.TemplateValues, error) {

	finalValues := models.TemplateValues{}

	for _, varsFilePath := range filePaths {
		file, err := os.Open(varsFilePath)
		if err != nil {
			return models.TemplateValues{}, err
		}

		// Read values from vars file
		lv := models.TemplateValues{}
		err = templator.ParseValues(&lv, file)
		if err != nil {
			return models.TemplateValues{}, err
		}

		finalValues = templator.MergeValues(finalValues, lv)

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
