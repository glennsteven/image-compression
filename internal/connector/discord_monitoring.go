package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"image-compressions/internal/config"
	"image-compressions/internal/request"
	"net/http"
	"time"
)

type Alerting interface {
	SendAlert(ctx context.Context, msg string) error
}

type DiscordAlerting struct {
	config config.Discord
	logger *logrus.Logger
	client *http.Client
}

func NewAlertingDiscord(
	cfg config.Discord,
	logger *logrus.Logger,
	client *http.Client,
) *DiscordAlerting {
	return &DiscordAlerting{
		config: cfg,
		logger: logger,
		client: client,
	}
}

func (d *DiscordAlerting) SendAlert(ctx context.Context, msg string) error {
	payload := request.DiscordRequest{Content: msg}

	req, err := json.Marshal(payload)
	if err != nil {
		d.logger.Printf("failed to marshal payload: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.config.Url, bytes.NewReader(req))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		d.logger.Printf("failed to send request to discord: %v", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}
