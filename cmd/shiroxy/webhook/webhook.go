// Author: @ShikharY10
// Docs Author: ChatGPT

// Package webhook provides utilities for handling webhook events, firing HTTP requests to
// specified URLs, and managing secrets for webhook authentication.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"sync"

	"crypto/rand"
	mathRand "math/rand"

	"github.com/google/uuid"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// WebhookFirePayload represents the payload structure sent to a webhook.
type WebhookFirePayload struct {
	EventName string      `json:"event_name"` // Name of the event triggering the webhook.
	Data      interface{} `json:"data"`       // Data payload associated with the event.
}

// WebhookHandler manages the configuration, logging, and execution of webhook events.
type WebhookHandler struct {
	logHandler    *logger.Logger           // Logger for logging webhook events and errors.
	WebHookConfig models.Webhook           // Webhook configuration, including the target URL and events.
	secret        string                   // Secret used for authenticating webhook requests.
	fire          chan *WebhookFirePayload // Channel to handle webhook payloads asynchronously.
}

// ApiResponse represents the structure of the response received from a webhook call.
type ApiResponse struct {
	Status  int    `json:"status"`  // Status code of the API response.
	Message string `json:"message"` // Message contained in the API response.
}

// StartWebhookHandler initializes the webhook handler with the specified configuration and secret.
// Parameters:
//   - config: models.Webhook, webhook configuration details (URL and events).
//   - logHandler: *logger.Logger, logger for logging events and errors.
//   - wg: *sync.WaitGroup, used for synchronization of goroutines.
//   - secret: string, the secret key used for webhook authentication (generated if not provided).
//
// Returns:
//   - *WebhookHandler: an instance of WebhookHandler.
//   - error: error if the secret generation fails.
func StartWebhookHandler(config models.Webhook, logHandler *logger.Logger, wg *sync.WaitGroup, secret string) (*WebhookHandler, error) {
	var selectedSecret string
	var err error

	// Generate a secret if not provided.
	if secret == "" {
		selectedSecret, err = generateSecret(10)
		if err != nil {
			return nil, err
		}
	} else {
		selectedSecret = secret
	}

	// Initialize the WebhookHandler.
	webhookHandler := &WebhookHandler{
		WebHookConfig: config,
		secret:        selectedSecret,
		fire:          make(chan *WebhookFirePayload, 1), // Channel for handling webhook events.
	}

	// Start a goroutine to process webhook events.
	wg.Add(1)
	go func() {
		defer wg.Done()
		payload := <-webhookHandler.fire
		webhookHandler.fireWebhook(payload)
	}()

	return webhookHandler, nil
}

// Fire triggers a webhook event with the specified event name and data.
// Parameters:
//   - eventName: string, the name of the event to fire.
//   - data: interface{}, the data associated with the event.
func (w *WebhookHandler) Fire(eventName string, data interface{}) {
	w.fire <- &WebhookFirePayload{
		EventName: eventName,
		Data:      data,
	}
}

// fireWebhook sends the webhook payload to the configured URL using an HTTP POST request.
// Parameters:
//   - payload: *WebhookFirePayload, the payload to be sent to the webhook.
func (w *WebhookHandler) fireWebhook(payload *WebhookFirePayload) {
	var eventFound bool

	// Check if the event is configured to be handled by the webhook.
	for _, event := range w.WebHookConfig.Events {
		if event == payload.EventName {
			eventFound = true
			break
		}
	}

	// Proceed only if the event is found in the configuration.
	if eventFound {
		// Marshal the payload data to JSON.
		jsonData, err := json.Marshal(payload.Data)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		// Create a new HTTP POST request to the webhook URL.
		req, err := http.NewRequest("POST", w.WebHookConfig.Url, bytes.NewBuffer(jsonData))
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		defer resp.Body.Close()

		// Read the response body from the webhook request.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		// Unmarshal the response into ApiResponse struct.
		var apiResponse ApiResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		// Log the outcome of the webhook call based on the status.
		if apiResponse.Status == 200 {
			w.logHandler.Log("webhook called successfully", "Webhook", "")
		} else {
			w.logHandler.Log("webhook failed!", "Webhook", "")
		}
	}
}

// generateSecret creates a random secret of the specified length.
// Parameters:
//   - length: int, the length of the secret to be generated.
//
// Returns:
//   - string: the generated secret string.
//   - error: error if the secret generation fails.
func generateSecret(length int) (string, error) {
	newUuid := uuid.New()
	var resultedString string

	// Generate a random string of the specified length using the letters constant.
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		resultedString += string(letters[index.Int64()])
	}

	// Generate a random numeric string of the specified length.
	var resultedNumber string
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(8))
		if err != nil {
			return "", err
		}
		resultedNumber += num.String()
	}

	// Combine the UUID, random number, and random string to form the secret.
	joined := fmt.Sprintf("%s%s%s", newUuid.String(), resultedNumber, resultedString)
	runes := []rune(joined)
	mathRand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})

	return string(runes), nil
}
