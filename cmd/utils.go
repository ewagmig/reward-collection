package cmd

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func PrintJSON(data interface{}) {
	r, _ := StructToJSON(data)
	fmt.Print(r)
}

func YAMLToStruct(yamlPath string, result interface{}) error {
	content, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, result)
}

func StructToYAML(s interface{}) (string, error) {
	r, err := yaml.Marshal(s)
	return string(r), err
}

func StructToJSON(s interface{}) (string, error) {
	r, err := json.Marshal(s)
	return string(r), err
}

