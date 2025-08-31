package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type DiscordWebhook struct {
	webhookURL string
	client     *http.Client
}

// WebhookMessage represents a Discord webhook message payload
type WebhookMessage struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

type Embed struct {
	Title  string       `json:"title,omitempty"`
	Fields []EmbedField `json:"fields,omitempty"`
	Image  EmbedImage   `json:"image,omitempty"`
	URL    string       `json:"url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type EmbedImage struct {
	URL string `json:"url"`
}

// NewDiscordWebhook creates a new Discord webhook client with sensible defaults
func NewDiscordWebhook(webhookURL string) *DiscordWebhook {
	return &DiscordWebhook{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewDiscordWebhookWithClient creates a new Discord webhook client with a custom HTTP client
func NewDiscordWebhookWithClient(webhookURL string, client *http.Client) *DiscordWebhook {
	return &DiscordWebhook{
		webhookURL: webhookURL,
		client:     client,
	}
}

func (d DiscordWebhook) PostMessage(ctx context.Context, message WebhookMessage) error {
	if message.Content == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	body, err := json.Marshal(message)
	if err != nil {
		slog.Error("failed to marshal webhook message", "error", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("failed to create HTTP request", "error", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		slog.Error("error posting to discord", "error", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.Info("message sent successfully to Discord", "status_code", resp.StatusCode)
		return nil
	}

	// Read response body for error details
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		slog.Error("failed to send message to Discord and failed to read error response",
			"status_code", resp.StatusCode, "read_error", readErr)
		return fmt.Errorf("failed to send message (status: %d) and failed to read error response: %w",
			resp.StatusCode, readErr)
	}

	slog.Error("failed to send message to Discord",
		"status_code", resp.StatusCode,
		"response_body", string(respBody))

	return fmt.Errorf("failed to send message (status: %d): %s", resp.StatusCode, string(respBody))
}
