package services

import (
	"alle/internal/models"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigurator(t *testing.T) {

	templator := NewTemplator()
	configurator := NewConfigurator(templator)

	t.Run("nginx_configurator_parse_test", func(t *testing.T) {
		config := `
releases:
  - name: nginx
    packages:
      - name: nginx
        path: ./testdata/configurator
        manifests:
          - name: deployment.yaml
            vars:
              - "./testdata/configurator/deployment_values.yaml"
          - name: service.yaml
            vars:
              - "./testdata/configurator/service_values.yaml"
        vars:
          - "./testdata/configurator/base_values.yaml"
        labels:
          pkg: nginx
`
		alleConfig := new(models.AlleConfig)
		configReader := bytes.NewBufferString(config)
		err := configurator.ParseConfig(alleConfig, "test_environment", configReader)
		assert.Nil(t, err)
		assert.Equal(t, "test_environment", alleConfig.Environment)
		assert.Equal(t, 1, len(alleConfig.Releases))
		assert.Equal(t, 1, len(alleConfig.Releases[0].Packages))
		assert.Equal(t, 1, len(alleConfig.Releases[0].Packages[0].VarsFilePaths))
		assert.Equal(t, "./testdata/configurator/base_values.yaml", alleConfig.Releases[0].Packages[0].VarsFilePaths[0])
		assert.Equal(t, 2, len(alleConfig.Releases[0].Packages[0].Manifests))

		assert.Equal(t, "deployment.yaml", alleConfig.Releases[0].Packages[0].Manifests[0].Name)
		assert.Equal(t, "./testdata/configurator/deployment_values.yaml", alleConfig.Releases[0].Packages[0].Manifests[0].VarsFilePaths[0])
		assert.Equal(t, "nginx-deployment.yaml", alleConfig.Releases[0].Packages[0].Manifests[0].GetFullName())
		assert.Equal(t, "./testdata/configurator/manifests/deployment.yaml", alleConfig.Releases[0].Packages[0].Manifests[0].GetTemplatePath())

		assert.Equal(t, "service.yaml", alleConfig.Releases[0].Packages[0].Manifests[1].Name)
		assert.Equal(t, "./testdata/configurator/service_values.yaml", alleConfig.Releases[0].Packages[0].Manifests[1].VarsFilePaths[0])
		assert.Equal(t, "nginx-service.yaml", alleConfig.Releases[0].Packages[0].Manifests[1].GetFullName())
		assert.Equal(t, "./testdata/configurator/manifests/service.yaml", alleConfig.Releases[0].Packages[0].Manifests[1].GetTemplatePath())
	})

	t.Run("nginx_configurator_labels_select_test", func(t *testing.T) {
		config := `
releases:
  - name: nginx
    packages:
      - name: nginx
        path: ./testdata/configurator
        manifests:
          - name: deployment.yaml
            vars:
              - "./testdata/configurator/deployment_values.yaml"
          - name: service.yaml
            vars:
              - "./testdata/configurator/service_values.yaml"
        vars:
          - "./testdata/configurator/base_values.yaml"
        labels:
          pkg: nginx
`
		alleConfig := new(models.AlleConfig)
		configReader := bytes.NewBufferString(config)
		err := configurator.ParseConfig(alleConfig, "test_environment", configReader)
		assert.Nil(t, err)

		packs := configurator.GetPackagesByLabels(alleConfig, []string{"pkg=nginx"})
		assert.Equal(t, 1, len(packs))

		packs = configurator.GetPackagesByLabels(alleConfig, []string{"pkg=unknown"})
		assert.Equal(t, 0, len(packs))

		packs = configurator.GetPackagesByLabels(alleConfig, []string{"bad_label"})
		assert.Equal(t, 0, len(packs))
	})

}
