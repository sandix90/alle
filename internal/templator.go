package internal

//
//import (
//	"bytes"
//	"fmt"
//	log "github.com/sirupsen/logrus"
//	"gopkg.in/yaml.v2"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//	"text/template"
//)
//
//type TemplateValues struct {
//	Values *map[string]interface{}
//}
//
//func MakeTemplate(templatePath string, b *TemplateValues) error {
//	tmpl := template.Must(template.ParseFiles(templatePath))
//
//	if tmpl != nil {
//		err := tmpl.Execute(os.Stdout, b)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func ParseAlleValues(values *TemplateValues, file string) error {
//	filename, err := filepath.Abs(file)
//	if err != nil {
//		log.Error("Read file error")
//	}
//	yamlFile, err := ioutil.ReadFile(filename)
//	if err != nil {
//		log.Errorf("Can't read file %s", filename)
//		return err
//	}
//	err = yaml.Unmarshal(yamlFile, &values.Values)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func RenderTemplate(tmplFilePath string, buffer *bytes.Buffer, values *TemplateValues) error {
//	tmpl := template.Must(template.ParseFiles(tmplFilePath))
//	err := tmpl.Execute(buffer, values)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func GetPackageStringManifests(pack *Package) ([]string, error) {
//
//	var templatesOutput []string
//
//	for _, manifest := range pack.Schema.Manifests {
//
//		file := filepath.Join(pack.Schema.Path, "manifests", manifest.Name)
//		err := Exists(file)
//		if err != nil {
//			return nil, err
//		}
//
//		finalValues := TemplateValues{}
//		for _, varFile := range pack.VarsFilePaths {
//			localValues := TemplateValues{}
//
//			err := ParseAlleValues(&localValues, varFile)
//			if err != nil {
//				return nil, err
//			}
//			if finalValues.Values != nil {
//				*finalValues.Values = MergeMaps(*finalValues.Values, *localValues.Values)
//			} else {
//				finalValues = localValues
//			}
//		}
//
//		for _, manifestVarFile := range manifest.VarsFilePaths {
//			localValues := TemplateValues{}
//
//			err := ParseAlleValues(&localValues, manifestVarFile)
//			if err != nil {
//				return nil, err
//			}
//			if finalValues.Values != nil {
//				*finalValues.Values = MergeMaps(*finalValues.Values, *localValues.Values)
//			} else {
//				finalValues = localValues
//			}
//		}
//
//		var tmplBytes bytes.Buffer
//		err = RenderTemplate(file, &tmplBytes, &finalValues)
//		if err != nil {
//			return nil, err
//		}
//		templatesOutput = append(templatesOutput, tmplBytes.String())
//	}
//	return templatesOutput, nil
//}
//
//func GetStringTemplatesByLabels(filepath string, environment string, labels []string) ([]string, error) {
//	workDir, err := os.Getwd()
//	log.Debugf("Workdir: %s", workDir)
//	log.Debugf("Using alle file: %s", filepath)
//	err = Exists(filepath)
//	if err != nil {
//		return nil, err
//	}
//	aleConfig := &AlleConfig{}
//	aleConfig.Environment = environment
//
//	err = UnmarshalAlleConfig(aleConfig, filepath)
//	if err != nil {
//		log.Error("Bad alle config")
//		return nil, err
//	}
//	var tmpls []string
//	for _, release := range aleConfig.Releases {
//
//		out, err := release.GetStringPreConfigManifests()
//		log.Debugln(out)
//
//		for _, pack := range release.Packages {
//
//			labelFound := FindByLabel(pack, labels)
//			if labelFound {
//				//tmpls, err = GetPackageStringManifests(pack)
//				tmpls, err = pack.GetStringSchemaManifests()
//				if err != nil {
//					return nil, err
//				}
//				fmt.Println(tmpls)
//			}
//		}
//	}
//	return tmpls, nil
//}
