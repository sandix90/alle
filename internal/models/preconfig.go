package models

type PreConfig struct {
	Name     string    `validate:"required"`
	Manifest *Manifest `validate:"dive" yaml:"manifest"`
	Secrets  string
	Order    int
}
