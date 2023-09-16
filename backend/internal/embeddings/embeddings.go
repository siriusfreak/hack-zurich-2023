package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

type PredictRequest struct {
	Instances []Instance `json:"instances"`
}

type Instance struct {
	Text string `json:"text"`
}

type PredictResponse struct {
	Predictions     []Prediction `json:"predictions"`
	DeployedModelID string       `json:"deployedModelId"`
}

type Prediction struct {
	TextEmbedding []float64 `json:"textEmbedding"`
}

func GetAccessToken() (string, error) {
	out, err := exec.Command("gcloud", "auth", "print-access-token").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func MakePredictionRequest(projectID string, request PredictRequest) (*PredictResponse, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://us-central1-aiplatform.googleapis.com/v1/projects/%s/locations/us-central1/publishers/google/models/multimodalembedding@001:predict", projectID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response PredictResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
