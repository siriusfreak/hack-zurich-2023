package templater

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"gopkg.in/yaml.v3"
)

type InitQuestionData struct {
	Language  string
	Question  string
	Documents []Document
}

type Document struct {
	Url     string
	Offset  int
	Content string
}

type Templater struct {
	InitQuestion string `yaml:"initQuestion"`
}

func New(configFile string) (*Templater, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return nil, err
	}

	var templater Templater
	err = yaml.Unmarshal(data, &templater)
	if err != nil {
		fmt.Println("Error parsing YAML file:", err)
		return nil, err
	}

	return &templater, nil
}

func (t *Templater) ProcessTemplateInitQuestionData(data []InitQuestionData) (string, error) {
	tmpl, err := template.New("questionTemplate").Parse(t.InitQuestion)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	for _, d := range data {
		err = tmpl.Execute(&output, d)
		if err != nil {
			return "", err
		}
	}

	return output.String(), nil
}
