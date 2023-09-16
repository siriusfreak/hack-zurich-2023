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

type Corner struct {
	Name     string `yaml:"name"`
	Question string `yaml:"question"`
	Answer   string `yaml:"answer"`
}

type Templater struct {
	InitQuestion string    `yaml:"initQuestion"`
	AllQuestions string    `yaml:"allQuestions"`
	Corners      []*Corner `yaml:"corners"`
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

func (t *Templater) ProcessTemplateAllQuestionsData(question string,
	language string) (string, error) {
	tmpl, err := template.New("questionTemplate").Parse(t.AllQuestions)
	if err != nil {
		return "", err
	}

	data := struct {
		Question string
		Language string
	}{
		Question: question,
		Language: language,
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return "", err
	}

	return output.String(), nil

}

func (t *Templater) GetCornerNames() []string {
	var names []string
	for _, c := range t.Corners {
		names = append(names, c.Name)
	}
	return names
}

func (t *Templater) GetCornerQuestion(cornerName string, question string) (string, error) {
	var corner *Corner
	for _, c := range t.Corners {
		if c.Name == cornerName {
			corner = c
		}
	}

	if corner == nil {
		return "", fmt.Errorf("corner not found")
	}

	tmpl, err := template.New(cornerName).Parse(corner.Question)
	if err != nil {
		return "", err
	}

	data := struct {
		Question string
	}{
		Question: question,
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (t *Templater) GetCornerResponse(cornerName string) (string, error) {
	for _, c := range t.Corners {
		if c.Name == cornerName {
			return c.Answer, nil
		}
	}

	return "", fmt.Errorf("corner not found")
}
