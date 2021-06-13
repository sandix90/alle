package models

type Wait struct {
	For       string `validate:"required"`
	Condition string `validate:"required"`
	Timeout   int
}

type TemplateValues struct {
	Values map[string]interface{}
}

type AlleConfig struct {
	Environment     string       `validate:"required"`
	Releases        []*Release   `yaml:"releases" validate:"dive"`
	GlobalPreConfig []*PreConfig `yaml:"pre_config" validate:"omitempty,dive"`
}
