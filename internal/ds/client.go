// Package ds provides a client for communicating with the DiskStation API.
package ds

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client holds the configuration and HTTP client for DiskStation API requests.
type Client struct {
	baseURL    string
	username   string
	password   string
	sid        string
	httpClient *http.Client
}

// apiResponse is the generic wrapper returned by the DiskStation API.
type apiResponse struct {
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
	Error   *apiError       `json:"error,omitempty"`
}

// apiError represents an error returned by the DiskStation API.
type apiError struct {
	Code int `json:"code"`
}

// loginData holds the session ID returned on successful authentication.
type loginData struct {
	Sid string `json:"sid"`
}

// NewClient creates a new DiskStation API client.
// If skipTLSVerify is true, TLS certificate verification is disabled.
func NewClient(host string, port int, username, password string, skipTLSVerify bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify, //nolint:gosec
		},
	}
	return &Client{
		baseURL:  fmt.Sprintf("http://%s:%d/webapi", host, port),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Login authenticates against the DiskStation API and stores the session ID.
func (c *Client) Login() error {
	params := url.Values{}
	params.Set("api", "SYNO.API.Auth")
	params.Set("version", "3")
	params.Set("method", "login")
	params.Set("account", c.username)
	params.Set("passwd", c.password)
	params.Set("session", "ds2api")
	params.Set("format", "sid")

	var data loginData
	if err := c.get("auth.cgi", params, &data); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	c.sid = data.Sid
	return nil
}

// Logout invalidates the current session on the DiskStation.
func (c *Client) Logout() error {
	params := url.Values{}
	params.Set("api", "SYNO.API.Auth")
	params.Set("version", "1")
	params.Set("method", "logout")
	params.Set("session", "ds2api")

	if err := c.get("auth.cgi", params, nil); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	return nil
}

// get performs a GET request to the given CGI endpoint with the provided query
// parameters. If out is non-nil, the response data field is unmarshalled into it.
func (c *Client) get(cgi string, params url.Values, out interface{}) error {
	if c.sid != "" {
		params.Set("_sid", c.sid)
	}

	rawURL := fmt.Sprintf("%s/%s?%s", c.baseURL, cgi, params.Encode())

	resp, err := c.httpClient.Get(rawURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	var ar apiResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return fmt.Errorf("unmarshalling response: %w", err)
	}

	if !ar.Success {
		code := 0
		if ar.Error != nil {
			code = ar.Error.Code
		}
		return fmt.Errorf("api returned error code %d", code)
	}

	if out != nil && ar.Data != nil {
		if err := json.Unmarshal(ar.Data, out); err != nil {
			return fmt.Errorf("unmarshalling data: %w", err)
		}
	}

	return nil
}
