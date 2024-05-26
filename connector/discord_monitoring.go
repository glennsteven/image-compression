package connector

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"image-compressions/request"
	"net/http"
)

func LogDiscord(logger *logrus.Logger, url string, payload request.DiscordRequest) error {
	req, err := json.Marshal(payload)
	if err != nil {
		logger.Printf("failed to marshal payload: %v", err)
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		logger.Printf("failed to send request to Discord: %v", err)
		return err
	}

	defer resp.Body.Close()

	return nil
}
