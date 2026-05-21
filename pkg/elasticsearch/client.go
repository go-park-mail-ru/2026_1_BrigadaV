package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	return respData, resp.StatusCode, nil
}

func (c *Client) Index(ctx context.Context, index, id string, doc any) error {
	path := fmt.Sprintf("/%s/_doc/%s", index, id)
	_, status, err := c.do(ctx, http.MethodPut, path, doc)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("elasticsearch index error: status %d", status)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, index, id string) error {
	path := fmt.Sprintf("/%s/_doc/%s", index, id)
	_, status, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	if status >= 400 && status != 404 {
		return fmt.Errorf("elasticsearch delete error: status %d", status)
	}
	return nil
}

type SearchRequest struct {
	Query any `json:"query"`
	Size  int `json:"size,omitempty"`
}

type SearchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source json.RawMessage `json:"_source"`
			Score  float64         `json:"_score"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *Client) Search(ctx context.Context, index string, req SearchRequest) (*SearchResponse, error) {
	path := fmt.Sprintf("/%s/_search", index)
	data, status, err := c.do(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("elasticsearch search error: status %d, body: %s", status, data)
	}

	var result SearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal search response: %w", err)
	}
	return &result, nil
}

func (c *Client) EnsureIndex(ctx context.Context, index string, mapping any) error {
	path := fmt.Sprintf("/%s", index)
	_, status, err := c.do(ctx, http.MethodHead, path, nil)
	if err != nil {
		return err
	}
	if status == 200 {
		return nil
	}
	_, status, err = c.do(ctx, http.MethodPut, path, mapping)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("elasticsearch create index error: status %d", status)
	}
	return nil
}
