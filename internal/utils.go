package internal

import (
	"errors"
	"gopkg.in/go-playground/validator.v9"
	"os"
)

var (
	validate = validator.New()
)

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		if m == nil{continue}
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func Exists(path string) error {
	if _, err := os.Stat(path); err != nil{
		if os.IsNotExist(err){
			//log.Errorf("File %s doesn't exist", path)
			return err
		} else {
			err := errors.New("Bad file.")
			return err
		}
	}
	return nil
}

func ValidateStruct(s interface{}) error{
	return validate.Struct(s)
}