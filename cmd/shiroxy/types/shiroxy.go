package types

import (
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"sync"
)

type APIContext struct {
	WaitGroup        *sync.WaitGroup
	WebhookHandler   *webhook.WebhookHandler
	HealthChecker    *proxy.HealthChecker
	DomainStorage    *domains.Storage
	AnalyticsHandler *analytics.AnalyticsConfiguration
	LogHandler       *logger.Logger
	LoadBalancer     *proxy.LoadBalancer
	Configuration    *models.Config
}
