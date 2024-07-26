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

// "domain-register-success"
// "domain-register-failed"
// "domain-ssl-success"
// "domain-ssl-failed"
// "domain-remove-success"
// "domain-remove-failed"
// "domain-update-success"
// "domain-update-failed"

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type WebhookFirePayload struct {
	event_name string
	data       interface{}
}

type WebhookHandler struct {
	logHandler    *logger.Logger
	WebHookConfig models.Webhook
	secret        string
	fire          chan *WebhookFirePayload
}

type ApiResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func StartWebhookHandler(config models.Webhook, logHandler *logger.Logger, wg *sync.WaitGroup, secret string) (*WebhookHandler, error) {

	var selectedSecret string
	var err error
	if secret == "" {
		selectedSecret, err = generateSecret()
		if err != nil {
			return nil, err
		}
	} else {
		selectedSecret = secret
	}

	webhookHandler := &WebhookHandler{
		WebHookConfig: config,
		secret:        selectedSecret,
		fire:          make(chan *WebhookFirePayload),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		payload := <-webhookHandler.fire
		webhookHandler.fireWebhook(payload)
	}()

	return webhookHandler, nil
}

func (w *WebhookHandler) Fire(eventName string, data interface{}) {
	w.fire <- &WebhookFirePayload{
		event_name: eventName,
		data:       data,
	}
}

func (w *WebhookHandler) fireWebhook(payload *WebhookFirePayload) {
	var eventFound bool

	for _, event := range w.WebHookConfig.Events {
		if event == payload.event_name {
			eventFound = true
			break
		}
	}

	if eventFound {
		jsonData, err := json.Marshal(payload.data)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

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

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		var apiResponse ApiResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			w.logHandler.LogError(err.Error(), "Webhook", "Error")
		}

		if apiResponse.Status == 200 {
			w.logHandler.Log("webhook called successfully", "Webhook", "")
		} else {
			w.logHandler.Log("webhook failed!", "Webhook", "")
		}
	}
}

// Function for generating new webhook secret
func generateSecret() (string, error) {
	newUuid := uuid.New()
	length := 8
	max := big.NewInt(8)

	var resultedString string
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		resultedString += string(letters[index.Int64()])
	}

	var resultedNumber string
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		resultedNumber += num.String()
	}

	joined := fmt.Sprintf("%s%s%s", newUuid.String(), resultedNumber, resultedString)
	runes := []rune(joined)
	mathRand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})

	return string(runes), nil
}
