package shutdown

import (
	"encoding/base64"
	"fmt"
	"os"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"

	"google.golang.org/protobuf/proto"
)

func LoadShutdownPersistence(logHandler logger.Logger, configuration *models.Config, storage *domains.Storage, analyticsConfiguration *analytics.AnalyticsConfiguration) {
	fileContent, err := os.ReadFile(fmt.Sprintf("%s/persistence.shiroxy", configuration.Default.DataPersistancePath))
	if err != nil {
		logHandler.LogError(err.Error(), "ShutDown", "Load Persistence S1")
		return
	}

	base64DecodedData, err := base64.StdEncoding.DecodeString(string(fileContent))
	if err != nil {
		logHandler.LogError(err.Error(), "ShutDown", "Load Persistence S2")
		return
	}

	var shutDown ShutdownMetadata
	err = proto.Unmarshal(base64DecodedData, &shutDown)
	if err != nil {
		logHandler.LogError(err.Error(), "ShutDown", "Load Persistence S3")
		return
	}

	var domainDataPersistence domains.DataPersistance
	err = proto.Unmarshal([]byte(shutDown.DomainMetadata), &domainDataPersistence)
	if err != nil {
		logHandler.LogError(err.Error(), "ShutDown", "Load Persistence S4")
		return
	}

	for _, domainMetadata := range domainDataPersistence.Domains {
		storage.DomainMetadata[domainMetadata.Domain] = domainMetadata
	}

	storage.WebhookSecret = shutDown.WebhookSecret
	logHandler.LogSuccess(fmt.Sprintf("Total %d Retrieved\n", len(domainDataPersistence.Domains)), "", "")
}
