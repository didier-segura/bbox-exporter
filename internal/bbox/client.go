package bbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	loginRoute     = "/api/v1/login"
	logoutRoute    = "/api/v1/logout"
	committedBug   = "committedas\":\n\t}"
	committedFix   = "committedas\":0\n\t}"
	defaultTimeout = 10 * time.Second
)

type Client struct {
	baseURL    string
	password   string
	httpClient *http.Client
}

func NewClient(baseURL, password string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("init cookie jar: %w", err)
	}

	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		password: password,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: defaultTimeout,
		},
	}, nil
}

func (c *Client) Login(ctx context.Context) error {
	form := url.Values{}
	form.Set("password", c.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(loginRoute), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("login failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) Logout(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(logoutRoute), http.NoBody)
	if err != nil {
		return fmt.Errorf("create logout request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("logout request: %w", err)
	}
	resp.Body.Close()
	return nil
}

func (c *Client) FetchCPU(ctx context.Context) (DeviceCPU, error) {
	return fetchSingle[DeviceCPU](ctx, c, "/api/v1/device/cpu")
}

func (c *Client) FetchMem(ctx context.Context) (DeviceMem, error) {
	return fetchSingle[DeviceMem](ctx, c, "/api/v1/device/mem")
}

func (c *Client) FetchWanIPStats(ctx context.Context) (WanIPStats, error) {
	return fetchSingle[WanIPStats](ctx, c, "/api/v1/wan/ip/stats")
}

func (c *Client) FetchWanIPInfo(ctx context.Context) (WanIPInfo, error) {
	return fetchSingle[WanIPInfo](ctx, c, "/api/v1/wan/ip")
}

func (c *Client) FetchLanStats(ctx context.Context) (LanStats, error) {
	return fetchSingle[LanStats](ctx, c, "/api/v1/lan/stats")
}

func (c *Client) FetchWireless24Stats(ctx context.Context) (WirelessStats, error) {
	return fetchSingle[WirelessStats](ctx, c, "/api/v1/wireless/24/stats")
}

func (c *Client) FetchWireless5Stats(ctx context.Context) (WirelessStats, error) {
	return fetchSingle[WirelessStats](ctx, c, "/api/v1/wireless/5/stats")
}

func fetchSingle[T any](ctx context.Context, c *Client, route string) (T, error) {
	var zero T

	body, err := c.get(ctx, route)
	if err != nil {
		return zero, err
	}

	body = bytes.ReplaceAll(body, []byte(committedBug), []byte(committedFix))

	var payload []T
	if err := json.Unmarshal(body, &payload); err != nil {
		return zero, fmt.Errorf("decode response for %s: %w", route, err)
	}
	if len(payload) == 0 {
		return zero, fmt.Errorf("empty response for %s", route)
	}

	return payload[0], nil
}

func (c *Client) get(ctx context.Context, route string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(route), nil)
	if err != nil {
		return nil, fmt.Errorf("build request for %s: %w", route, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", route, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("status %d for %s body=%s", resp.StatusCode, route, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", route, err)
	}

	return body, nil
}

func (c *Client) url(route string) string {
	if strings.HasPrefix(route, "/") {
		return c.baseURL + route
	}
	return c.baseURL + "/" + route
}
