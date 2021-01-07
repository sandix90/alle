package services

import (
	"alle/internal"
	"alle/internal/models"
	"bytes"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"io/ioutil"
)

type Templator interface {
	ParseValues(values *models.TemplateValues, r io.Reader) error
	RenderTemplate(template io.Reader, buffer io.Writer, values models.TemplateValues) error
	MergeValues(fst models.TemplateValues, sec models.TemplateValues) models.TemplateValues
}

type TemplatorImpl struct{}

func NewTemplator() *TemplatorImpl {
	return &TemplatorImpl{}
}

func (t *TemplatorImpl) ParseValues(values *models.TemplateValues, source io.Reader) error {

	readBytes, err := ioutil.ReadAll(source)
	if err != nil {
		log.Errorf("Can't read from reader")
		return err
	}
	err = yaml.Unmarshal(readBytes, &values.Values)
	if err != nil {
		return err
	}
	return nil
}

// Values from sec will override values from fst
func (t *TemplatorImpl) MergeValues(fst models.TemplateValues, sec models.TemplateValues) models.TemplateValues {
	newTemplateValues := models.TemplateValues{}

	mergedValues := internal.MergeMaps(fst.Values, sec.Values)
	newTemplateValues.Values = mergedValues
	return newTemplateValues

}

func (t *TemplatorImpl) RenderTemplate(tmplReader io.Reader, writer io.Writer, values models.TemplateValues) error {
	tmpl := template.New("alle_template")

	readerBuf := new(bytes.Buffer)
	_, err := readerBuf.ReadFrom(tmplReader)
	if err != nil {
		return err
	}

	tmpl, err = tmpl.Parse(readerBuf.String())
	if err != nil {
		return err
	}

	err = tmpl.Execute(writer, values)
	if err != nil {
		return err
	}

	return nil
}
