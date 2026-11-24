package httpclient

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

type HTTPWebhookSender struct {
	client *http.Client
}

func NewHTTPWebhookSender() *HTTPWebhookSender {
	return &HTTPWebhookSender{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *HTTPWebhookSender) Send(ctx context.Context, targetURL, secretKey, payload string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBufferString(payload))
	if err != nil {
		return 0, err
	}

	// Sign payload using HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature-256", signature)

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
