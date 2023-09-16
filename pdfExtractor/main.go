package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"pdfextractor/client"
	"pdfextractor/esclient"
)

const (
	offset          = 400
	symbolsPerBlock = 800
)

var projectId = "hackzurich23-8200"
var url = "https://hz.siriusfrk.me/sika_chat_index/_doc/"
var username = ""
var password = "" //

func main() {
	rootDirectory := "../"

	err := filepath.Walk(rootDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".pdf") {
			processPDFFile(path, info.Name())
			log.Printf("Path: %v\n", path)
			log.Printf("Name: %v\n", info.Name())
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking directory: %v\n", err)
	}
}

func processPDFFile(pdfPath, fileName string) {

	textBlocks, err := extractTextFromPDF(pdfPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error extracting text:", err)
		os.Exit(1)
	}

	i := 0

	for lineNum, block := range textBlocks {
		prediction, error := client.MakePredictionRequest(projectId, client.PredictRequest{
			Instances: []client.Instance{
				{
					Text: block,
				},
			},
		})
		if error != nil {
			fmt.Println("Error:", error)
			return
		}
		embed := prediction.Predictions

		md5, err := calculateMD5(fileName, i)

		if err != nil {
			fmt.Println("Error:", error)
			return
		}

		if prediction.Predictions == nil {
			fmt.Println("Error getting embeddings:", block)
			fmt.Println(prediction)
			return
		}

		currentTime := time.Now().Format(time.RFC3339Nano)
		fmt.Println(currentTime)

		response, err := esclient.IndexData(url+md5, username, password, esclient.IndexRequest{
			Content:   block,
			Links:     []string{pdfPath},
			Offset:    lineNum * offset,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
			Embedding: embed[len(embed)-1].TextEmbedding, // add more values to match the dimension specified in the index settings
		})
		i = i + 1
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("Response:", response)
	}
}

func pdfToText(path string) (string, error) {
	out, err := exec.Command("pdftotext", path, "-").Output()
	return string(out), err
}

func rotate(inp []string) {
	for i, l := offset, len(inp); i < l; i++ {
		inp[i-offset] = inp[i]
	}
}

func glue(inp []string) string {
	var sb strings.Builder
	for i, l := 0, len(inp); i < l; i++ {
		sb.WriteString(inp[i])
	}
	return sb.String()
}

func extractTextFromPDF(path string) (map[int]string, error) {
	text, err := pdfToText(path)
	if err != nil {
		return nil, fmt.Errorf("pdfToText: %w", err)
	}

	textBlocks := make(map[int]string)
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanRunes)
	window := make([]string, symbolsPerBlock)

	for i := 0; i < symbolsPerBlock && scanner.Scan(); i++ {
		window[i] = scanner.Text()
	}

	textBlocks[len(textBlocks)] = glue(window)
	rotate(window)

	windowCurrent := offset
	for scanner.Scan() {
		if windowCurrent == len(window) {
			windowCurrent = offset
			textBlocks[len(textBlocks)] = glue(window)
			rotate(window)
		}

		window[windowCurrent] = scanner.Text()
		windowCurrent++
	}

	for i := windowCurrent; i < len(window); i++ {
		window[i] = ""
	}
	textBlocks[len(textBlocks)] = glue(window)

	return textBlocks, nil
}

func calculateMD5(text string, number int) (string, error) {
	// Concatenate the text and the integer as a string
	input := fmt.Sprintf("%s%d", text, number)

	// Calculate the MD5 hash
	hasher := md5.New()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		return "", err
	}

	// Convert the hash to a hexadecimal string
	hashSum := hex.EncodeToString(hasher.Sum(nil))
	return hashSum, nil
}
