package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type DispatcherConfig struct {
	// Email settings
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// Webhook settings
	WebhookTimeout     time.Duration
	MaxRetries         int
	RetryInitialDelay  time.Duration

	// Worker settings
	WorkerCount int
	QueueSize   int
}

type Dispatcher struct {
	config     DispatcherConfig
	httpClient *http.Client
	queue      chan dispatchJob
	wg         sync.WaitGroup
	stopCh     chan struct{}
}

type dispatchJob struct {
	channel      domain.NotificationChannel
	notification domain.Notification
}

func NewDispatcher(config DispatcherConfig) *Dispatcher {
	if config.WorkerCount == 0 {
		config.WorkerCount = 3
	}
	if config.QueueSize == 0 {
		config.QueueSize = 100
	}
	if config.WebhookTimeout == 0 {
		config.WebhookTimeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryInitialDelay == 0 {
		config.RetryInitialDelay = 1 * time.Second
	}

	return &Dispatcher{
		config: config,
		httpClient: &http.Client{
			Timeout: config.WebhookTimeout,
		},
		queue:  make(chan dispatchJob, config.QueueSize),
		stopCh: make(chan struct{}),
	}
}

func (d *Dispatcher) Start() {
	for i := 0; i < d.config.WorkerCount; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}
	log.Printf("notification dispatcher started with %d workers", d.config.WorkerCount)
}

func (d *Dispatcher) Stop() {
	close(d.stopCh)
	d.wg.Wait()
	log.Printf("notification dispatcher stopped")
}

func (d *Dispatcher) worker(id int) {
	defer d.wg.Done()

	for {
		select {
		case <-d.stopCh:
			return
		case job := <-d.queue:
			if err := d.processJob(job); err != nil {
				log.Printf("failed to dispatch notification: channel_type=%s notification_type=%s error=%v",
					job.channel.ChannelType, job.notification.Type, err)
			}
		}
	}
}

func (d *Dispatcher) Dispatch(channel domain.NotificationChannel, notification domain.Notification) {
	select {
	case d.queue <- dispatchJob{channel: channel, notification: notification}:
	default:
		log.Printf("notification dispatch queue full, dropping notification")
	}
}

func (d *Dispatcher) DispatchSync(channel domain.NotificationChannel, notification domain.Notification) error {
	return d.processJob(dispatchJob{channel: channel, notification: notification})
}

func (d *Dispatcher) processJob(job dispatchJob) error {
	switch job.channel.ChannelType {
	case domain.ChannelTypeEmail:
		return d.sendEmail(job.channel, job.notification)
	case domain.ChannelTypeWebhook:
		return d.sendWebhook(job.channel, job.notification)
	default:
		return fmt.Errorf("unsupported channel type: %s", job.channel.ChannelType)
	}
}

func (d *Dispatcher) sendEmail(channel domain.NotificationChannel, notification domain.Notification) error {
	recipients := channel.GetEmailRecipients()
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients configured")
	}

	subject := notification.Title
	body := notification.Body

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		d.config.SMTPFrom,
		recipients[0],
		subject,
		body,
	)

	addr := fmt.Sprintf("%s:%d", d.config.SMTPHost, d.config.SMTPPort)

	var auth smtp.Auth
	if d.config.SMTPUser != "" {
		auth = smtp.PlainAuth("", d.config.SMTPUser, d.config.SMTPPassword, d.config.SMTPHost)
	}

	return d.sendWithRetry(func() error {
		return smtp.SendMail(addr, auth, d.config.SMTPFrom, recipients, []byte(msg))
	})
}

func (d *Dispatcher) sendWebhook(channel domain.NotificationChannel, notification domain.Notification) error {
	url := channel.GetWebhookURL()
	if url == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	payload := map[string]interface{}{
		"type":       notification.Type,
		"title":      notification.Title,
		"body":       notification.Body,
		"metadata":   notification.Metadata,
		"created_at": notification.CreatedAt.Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	secret := channel.GetWebhookSecret()

	return d.sendWithRetry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), d.config.WebhookTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Claudeception/1.0")

		if secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			signature := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-Signature-256", "sha256="+signature)
		}

		resp, err := d.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("webhook returned status %d", resp.StatusCode)
		}

		return nil
	})
}

func (d *Dispatcher) sendWithRetry(fn func() error) error {
	var lastErr error
	delay := d.config.RetryInitialDelay

	for attempt := 0; attempt < d.config.MaxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			log.Printf("dispatch attempt %d/%d failed: %v, retrying", attempt+1, d.config.MaxRetries, err)
			time.Sleep(delay)
			delay *= 2
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", d.config.MaxRetries, lastErr)
}
