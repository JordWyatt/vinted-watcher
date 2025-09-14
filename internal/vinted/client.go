package vinted

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"vinted-watcher/internal/domain"
)

const PROXIES_ENV_VAR = "PROXY_URLS"
const vintedAuthURL = "https://www.vinted.co.uk"

type VintedClient interface {
	GetItems(params *domain.SearchParams) ([]Item, error)
}

type Client struct {
	baseURL      string
	httpClient   *http.Client
	proxies      []url.URL
	currentProxy int
}

func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	client := Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Jar: jar,
		},
		proxies: getProxies(),
	}

	if len(client.proxies) > 0 {
		slog.Info("Using proxies", "proxies", client.proxies)
	} else {
		slog.Info("No proxies configured")
	}

	err := client.ResetSession()
	if err != nil {
		slog.Error("Error initializing Vinted client session, continuing anyway", "error", err)
	}
	return &client
}

// ResetSession discards cookies and re-initiates a session.
func (c *Client) ResetSession() error {
	jar, _ := cookiejar.New(nil)
	c.httpClient.Jar = jar

	req, _ := http.NewRequest(http.MethodGet, c.baseURL, nil)
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("reset session failed: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("Session reset", "status", resp.Status)
	return nil
}

func (c *Client) GetItems(params *domain.SearchParams) ([]Item, error) {
	apiURL, err := params.ToApiURL()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}
	slog.Info("Making Vinted API request", "vinted_api_url", req.URL.String())

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		slog.Warn("Got 401, re-initializing Vinted session")
		resp.Body.Close()

		if err := c.ResetSession(); err != nil {
			return nil, fmt.Errorf("failed to re-init session: %w", err)
		}

		resp, err = c.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var itemsResponse ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return itemsResponse.Items, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	setDefaultHeaders(req)

	// No proxies configured
	if len(c.proxies) == 0 {
		return c.httpClient.Do(req)
	}

	// Rotate to next proxy
	proxy := &c.proxies[c.currentProxy]
	c.currentProxy = (c.currentProxy + 1) % len(c.proxies)
	slog.Info("Using proxy", "proxy", proxy.String())

	// Update the existing httpClient transport to use the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	c.httpClient.Transport = transport

	return c.httpClient.Do(req)
}

func getProxies() []url.URL {
	commaSeparatedProxies := os.Getenv(PROXIES_ENV_VAR)
	if commaSeparatedProxies == "" {
		return nil
	}

	proxies := strings.Split(commaSeparatedProxies, ",")
	validProxies := make([]url.URL, 0)

	for i := range proxies {
		url, err := url.Parse(proxies[i])
		if err != nil {
			slog.Warn("Invalid proxy URL, skipping", "proxy_url", proxies[i], "error", err)
			continue
		}

		validProxies = append(validProxies, *url)
	}

	return validProxies
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.google.com/")
	req.Header.Set("Connection", "keep-alive")
}
