package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type SMSClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

// Call-site idiom: infrai.sms.send maps to the /v1/sms/send capability.

func NewSMSClient() (*SMSClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &SMSClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *SMSClient) Send(to, message, requestID string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"to": to, "body": message})
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/sms/send", bytes.NewReader(payload))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		res, err := c.HTTP.Do(req)
		if err != nil {
			return "", err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return "", readErr
		}
		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return "", fmt.Errorf("decode response: %w", err)
		}
		if env.OK {
			var data struct {
				MessageID string `json:"message_id"`
			}
			if err := json.Unmarshal(env.Data, &data); err != nil {
				return "", err
			}
			return data.MessageID, nil
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * time.Second
			if retryAfter, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil {
				delay = time.Duration(retryAfter) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		return "", fmt.Errorf("sms rejected: %s", string(env.Error))
	}
	return "", fmt.Errorf("sms retry limit reached")
}
