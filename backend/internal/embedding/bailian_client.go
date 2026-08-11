package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const defaultDimensions = 1024

type BailianClient struct {
	APIKey     string
	BaseURL    string
	Model      string
	Dimensions int
	Client     *http.Client
}

type embeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	Dimensions     int      `json:"dimensions"`
	EncodingFormat string   `json:"encoding_format"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func NewBailianClient(apiKey string, baseURL string) *BailianClient {
	return &BailianClient{
		APIKey:     apiKey,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Model:      "text-embedding-v4",
		Dimensions: defaultDimensions,
		Client:     &http.Client{},
	}
}

func (c *BailianClient) Embed(
	ctx context.Context,
	text string,
) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("embedding input must not be empty")
	}

	payload := embeddingRequest{
		Model:          c.Model,
		Input:          []string{text},
		Dimensions:     c.Dimensions,
		EncodingFormat: "float",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/embeddings",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"bailian returned status %s",
			resp.Status,
		)
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	if len(result.Data) != 1 {
		return nil, fmt.Errorf(
			"expected 1 embedding, got %d",
			len(result.Data),
		)
	}

	source := result.Data[0].Embedding

	if len(source) != c.Dimensions {
		return nil, fmt.Errorf(
			"expected embedding dimension %d, got %d",
			c.Dimensions,
			len(source),
		)
	}

	vector := make([]float32, len(source))
	for i, value := range source {
		vector[i] = float32(value)
	}

	return vector, nil
}
