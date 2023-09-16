package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SearchRequest struct {
	KNN struct {
		Field         string    `json:"field"`
		QueryVector   []float64 `json:"query_vector"`
		K             int       `json:"k"`
		NumCandidates int       `json:"num_candidates"`
	} `json:"knn"`
	Size int `json:"size"`
}

type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Content   string    `json:"content"`
				Links     []string  `json:"links"`
				CreatedAt string    `json:"created_at"`
				UpdatedAt string    `json:"updated_at"`
				Embedding []float64 `json:"embedding"`
				Offset    int       `json:"offset"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func Search(index string, request SearchRequest) (*SearchResponse, error) {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://hz.siriusfrk.me/"+index+"/_search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic ZWxhc3RpYzpRV0UhIzJhc2R6eGM=")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d. Body %s", resp.StatusCode, resp.Body)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResponse SearchResponse
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		return nil, err
	}

	return &searchResponse, nil
}
