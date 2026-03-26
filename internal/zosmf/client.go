package zosmf

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	logger "zcli/internal/logging"
)

// Client holds the connection details for a z/OSMF server.
type Client struct {
	BaseURL    string
	User       string
	Password   string
	Verify     bool
	Timeout    time.Duration
	httpClient *http.Client
}

// Response wraps the HTTP response with parsed body.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// APIError represents an error returned from a z/OSMF API call.
type APIError struct {
	RC         int    `json:"rc"`
	StatusCode int    `json:"status_code"`
	Reason     string `json:"reason"`
	Message    string `json:"message,omitempty"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("rc=%d, status=%d, reason=%s\n%s", e.RC, e.StatusCode, e.Reason, e.Message)
	}
	return fmt.Sprintf("rc=%d, status=%d, reason=%s", e.RC, e.StatusCode, e.Reason)
}

// NewClient creates a new z/OSMF client.
func NewClient(protocol, hostname, port, user, password string, verify bool) *Client {
	baseURL := fmt.Sprintf("%s://%s:%s/zosmf", protocol, hostname, port)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !verify,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &Client{
		BaseURL:  baseURL,
		User:     user,
		Password: password,
		Verify:   verify,
		Timeout:  30 * time.Second,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// basicAuth returns the Base64-encoded "user:password" string.
func (c *Client) basicAuth() string {
	auth := c.User + ":" + c.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// newRequest creates an HTTP request with z/OSMF standard headers.
func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-CSRF-ZOSMF-HEADER", "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+c.basicAuth())

	return req, nil
}

// doRequest executes an HTTP request and returns the parsed response.
func (c *Client) doRequest(req *http.Request) (*Response, error) {
	logger.Log.Debug("z/OSMF request",
		"method", req.Method,
		"url", req.URL.String(),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Log.Debug("z/OSMF response",
		"status", resp.StatusCode,
		"body_length", len(body),
	)

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

// Get performs a GET request to the given z/OSMF API path.
func (c *Client) Get(path string, headers map[string]string) (*Response, error) {
	url := c.BaseURL + path
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doRequest(req)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, payload interface{}, headers map[string]string) (*Response, error) {
	url := c.BaseURL + path
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}
	req, err := c.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doRequest(req)
}

// Put performs a PUT request with a JSON or raw body.
func (c *Client) Put(path string, payload interface{}, headers map[string]string) (*Response, error) {
	url := c.BaseURL + path
	var body io.Reader
	if payload != nil {
		switch v := payload.(type) {
		case []byte:
			body = bytes.NewReader(v)
		case string:
			body = bytes.NewReader([]byte(v))
		default:
			jsonData, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payload: %w", err)
			}
			body = bytes.NewReader(jsonData)
		}
	}
	req, err := c.newRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doRequest(req)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string, headers map[string]string) (*Response, error) {
	url := c.BaseURL + path
	req, err := c.newRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doRequest(req)
}

// PutRaw performs a PUT request with a raw byte body (for job submit etc.).
func (c *Client) PutRaw(path string, data []byte, headers map[string]string) (*Response, error) {
	url := c.BaseURL + path
	req, err := c.newRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doRequest(req)
}

// CheckResponse validates the HTTP status code and returns an APIError if unexpected.
func CheckResponse(resp *Response, expectedStatus ...int) error {
	for _, s := range expectedStatus {
		if resp.StatusCode == s {
			return nil
		}
	}
	return &APIError{
		RC:         8,
		StatusCode: resp.StatusCode,
		Reason:     fmt.Sprintf("unexpected status code %d", resp.StatusCode),
		Message:    string(resp.Body),
	}
}

// BodyString returns the response body as string.
func (r *Response) BodyString() string {
	return string(r.Body)
}

// JSON unmarshals the response body into the given target.
func (r *Response) JSON(target interface{}) error {
	return json.Unmarshal(r.Body, target)
}
