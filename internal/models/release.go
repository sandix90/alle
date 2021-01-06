package models

type Release struct {
	Name      string       `validate:"required"`
	Packages  []*Package   `validate:"dive"`
	PreConfig []*PreConfig `validate:"dive"`
}
