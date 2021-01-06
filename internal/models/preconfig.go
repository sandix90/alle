package models

type PreConfig struct {
	Name     string    `validate:"required"`
	Manifest *Manifest `validate:"dive"`
	Secrets  string
	Order    int
}
