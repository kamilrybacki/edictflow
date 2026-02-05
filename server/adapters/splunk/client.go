package splunk

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds Splunk HEC configuration
type Config struct {
	HECURL        string
	Token         string
	Source        string
	SourceType    string
	Index         string
	Timeout       time.Duration
	SkipTLSVerify bool
}

// Client wraps HTTP operations for Splunk HEC
type Client struct {
	httpClient *http.Client
	config     Config
}

// Event represents a Splunk HEC event
type Event struct {
	Time       float64                `json:"time,omitempty"`
	Host       string                 `json:"host,omitempty"`
	Source     string                 `json:"source,omitempty"`
	SourceType string                 `json:"sourcetype,omitempty"`
	Index      string                 `json:"index,omitempty"`
	Event      map[string]interface{} `json:"event"`
}

// NewClient creates a new Splunk HEC client
func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{}
	if cfg.SkipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		config: cfg,
	}
}

// Send posts a single event to Splunk HEC
func (c *Client) Send(ctx context.Context, event Event) error {
	c.applyDefaults(&event)

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return c.doRequest(ctx, body)
}

// SendBatch posts multiple events to Splunk HEC as newline-delimited JSON
func (c *Client) SendBatch(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for i := range events {
		c.applyDefaults(&events[i])
		eventBytes, err := json.Marshal(events[i])
		if err != nil {
			return fmt.Errorf("failed to marshal event %d: %w", i, err)
		}
		buf.Write(eventBytes)
		buf.WriteByte('\n')
	}

	return c.doRequest(ctx, buf.Bytes())
}

// Ping checks the Splunk HEC health
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.HECURL+"/services/collector/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	req.Header.Set("Authorization", "Splunk "+c.config.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ping failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) applyDefaults(event *Event) {
	if event.Time == 0 {
		event.Time = float64(time.Now().UnixMilli()) / 1000.0
	}
	if event.Source == "" {
		event.Source = c.config.Source
	}
	if event.SourceType == "" {
		event.SourceType = c.config.SourceType
	}
	if event.Index == "" {
		event.Index = c.config.Index
	}
}

func (c *Client) doRequest(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.HECURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Splunk "+c.config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HEC returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
