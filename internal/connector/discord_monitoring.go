package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"image-compressions/request"
	"net/http"
	"time"
)

func LogDiscord(logger *logrus.Logger, url string, payload request.DiscordRequest) error {
	req, err := json.Marshal(payload)
	if err != nil {
		logger.Printf("failed to marshal payload: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(req))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logger.Printf("failed to send request to discord: %v", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}
