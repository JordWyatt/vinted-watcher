package vinted

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"vinted-watcher/internal/domain"
)

const vintedAuthURL = "https://www.vinted.co.uk"

type VintedClient interface {
	GetItems(params *domain.SearchParams) ([]Item, error)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ===== Top-level API Response =====
type ItemsResponse struct {
	Code                 int                  `json:"code"`
	Pagination           Pagination           `json:"pagination"`
	SearchTrackingParams SearchTrackingParams `json:"search_tracking_params"`
	Items                []Item               `json:"items"`
}

// ===== Metadata =====
type Pagination struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalEntries int `json:"total_entries"`
	PerPage      int `json:"per_page"`
	Time         int `json:"time"`
}

type SearchTrackingParams struct {
	SearchCorrelationID   string `json:"search_correlation_id"`
	GlobalSearchSessionID string `json:"global_search_session_id"`
	SearchSessionID       string `json:"search_session_id"`
}

// ===== Item-level Data =====
type Item struct {
	ID                   int64                     `json:"id"`
	Title                string                    `json:"title"`
	Price                Price                     `json:"price"`
	IsVisible            bool                      `json:"is_visible"`
	BrandTitle           string                    `json:"brand_title"`
	Path                 string                    `json:"path"`
	User                 User                      `json:"user"`
	Conversion           any                       `json:"conversion"`
	URL                  string                    `json:"url"`
	Promoted             bool                      `json:"promoted"`
	Photo                ItemPhoto                 `json:"photo"`
	FavouriteCount       int                       `json:"favourite_count"`
	IsFavourite          bool                      `json:"is_favourite"`
	ServiceFee           ServiceFee                `json:"service_fee"`
	TotalItemPrice       TotalItemPrice            `json:"total_item_price"`
	ViewCount            int                       `json:"view_count"`
	SizeTitle            string                    `json:"size_title"`
	ContentSource        string                    `json:"content_source"`
	Status               string                    `json:"status"`
	ItemBox              ItemBox                   `json:"item_box,omitempty"`
	SearchTrackingParams ItemsSearchTrackingParams `json:"search_tracking_params"`
}

type ItemsSearchTrackingParams struct {
	Score          float64 `json:"score"`
	MatchedQueries []any   `json:"matched_queries"`
}

// ===== Price-related Structs =====
type Price struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

type ServiceFee struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

type TotalItemPrice struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

// ===== Item Visuals =====
type ItemPhoto struct {
	ID                  int            `json:"id"`
	Width               int            `json:"width"`
	Height              int            `json:"height"`
	TempUUID            any            `json:"temp_uuid"`
	URL                 string         `json:"url"`
	DominantColor       string         `json:"dominant_color"`
	DominantColorOpaque string         `json:"dominant_color_opaque"`
	Thumbnails          []Thumbnails   `json:"thumbnails"`
	IsSuspicious        bool           `json:"is_suspicious"`
	Orientation         any            `json:"orientation"`
	HighResolution      HighResolution `json:"high_resolution"`
	FullSizeURL         string         `json:"full_size_url"`
	IsHidden            bool           `json:"is_hidden"`
	Extra               Extra          `json:"extra"`
}

type Thumbnails struct {
	Type         string `json:"type"`
	URL          string `json:"url"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	OriginalSize any    `json:"original_size"`
}

type HighResolution struct {
	ID          string `json:"id"`
	Timestamp   int    `json:"timestamp"`
	Orientation any    `json:"orientation"`
}

type Extra struct{}

// ===== Item Additional Info =====
type ItemBox struct {
	FirstLine          string `json:"first_line"`
	SecondLine         string `json:"second_line"`
	Exposures          []any  `json:"exposures"`
	AccessibilityLabel string `json:"accessibility_label"`
	ItemID             int64  `json:"item_id"`
}

type Badge struct {
	Title string `json:"title"`
}

// ===== User Info =====
type User struct {
	ID         int       `json:"id"`
	Login      string    `json:"login"`
	ProfileURL string    `json:"profile_url"`
	Photo      UserPhoto `json:"photo"`
	Business   bool      `json:"business"`
}

type UserPhoto struct {
	ID                  int64          `json:"id"`
	ImageNo             int            `json:"image_no"`
	Width               int            `json:"width"`
	Height              int            `json:"height"`
	DominantColor       string         `json:"dominant_color"`
	DominantColorOpaque string         `json:"dominant_color_opaque"`
	URL                 string         `json:"url"`
	IsMain              bool           `json:"is_main"`
	Thumbnails          []Thumbnails   `json:"thumbnails"`
	HighResolution      HighResolution `json:"high_resolution"`
	IsSuspicious        bool           `json:"is_suspicious"`
	FullSizeURL         string         `json:"full_size_url"`
	IsHidden            bool           `json:"is_hidden"`
	Extra               Extra          `json:"extra"`
}

func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Jar: jar,
		},
	}
}

// InitSession initializes a session with the Vinted API. Cookies are stored in the client's cookie jar.
func (c *Client) InitSession() error {
	req, err := http.NewRequest(http.MethodHead, vintedAuthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create session request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate session: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) GetItems(params *domain.SearchParams) ([]Item, error) {
	if err := c.InitSession(); err != nil {
		return nil, err
	}

	apiURL, err := params.ToApiURL()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}

	slog.Info("Making Vinted API request", "vinted_api_url", req.URL.String())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var itemsResponse ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return itemsResponse.Items, nil
}
