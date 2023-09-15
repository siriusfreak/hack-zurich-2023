package pallm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
)

type RequestParameters struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
	TopK            int     `json:"topK"`
	TopP            float64 `json:"topP"`
}

type requestBody struct {
	Instances  []map[string]string `json:"instances"`
	Parameters RequestParameters   `json:"parameters"`
}

type SafetyAttributes struct {
	Scores     []float64 `json:"scores"`
	Categories []string  `json:"categories"`
	Blocked    bool      `json:"blocked"`
}

type CitationMetadata struct {
	Citations []interface{} `json:"citations"`
}

type Prediction struct {
	SafetyAttributes SafetyAttributes `json:"safetyAttributes"`
	CitationMetadata CitationMetadata `json:"citationMetadata"`
	Content          string           `json:"content"`
}

type TokenCount struct {
	TotalBillableCharacters int `json:"totalBillableCharacters"`
	TotalTokens             int `json:"totalTokens"`
}

type TokenMetadata struct {
	InputTokenCount  TokenCount `json:"inputTokenCount"`
	OutputTokenCount TokenCount `json:"outputTokenCount"`
}

type Metadata struct {
	TokenMetadata TokenMetadata `json:"tokenMetadata"`
}

type Response struct {
	Predictions []Prediction `json:"predictions"`
	Metadata    Metadata     `json:"metadata"`
}

func GetAccessToken() (string, error) {
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out[0 : len(out)-1]), nil
}

func MakeRequest(prompt string, params RequestParameters) (Response, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return Response{}, err
	}

	body := requestBody{
		Instances: []map[string]string{
			{"prompt": prompt},
		},
		Parameters: params,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", "https://us-central1-aiplatform.googleapis.com/v1/projects/hackzurich23-8200/locations/us-central1/publishers/google/models/text-bison:predict", bytes.NewBuffer(jsonBody))
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Authorization", "Bearer "+string(accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var parsed Response
	err = json.Unmarshal(respBody, &parsed)

	return parsed, err
}
