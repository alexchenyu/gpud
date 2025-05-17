package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type NotificationType string

const (
	NotificationTypeShutdown NotificationType = "shutdown"
	NotificationTypeStartup  NotificationType = "startup"
)

type payload struct {
	ID   string           `json:"id"`
	Type NotificationType `json:"type"`
}

func notification(endpoint string, req payload) error {
	type RespErr struct {
		Error  string `json:"error"`
		Status string `json:"status"`
	}
	rawPayload, _ := json.Marshal(&req)
	response, err := http.Post(createNotificationURL(endpoint), "application/json", bytes.NewBuffer(rawPayload))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		var errorResponse RespErr
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return fmt.Errorf("Error parsing error response: %v\nResponse body: %s", err, body)
		}
		return fmt.Errorf("failed to send notification: %v", errorResponse)
	}
	return nil
}

// createNotificationURL creates a URL for the notification endpoint
func createNotificationURL(endpoint string) string {
	host := endpoint
	url, _ := url.Parse(endpoint)
	if url.Host != "" {
		host = url.Host
	}
	return fmt.Sprintf("https://%s/api/v1/notification", host)
}
