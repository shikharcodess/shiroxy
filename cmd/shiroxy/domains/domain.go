// Package domains provides functionalities for managing domain metadata, including registration,
// updating, removal, and certificate generation for domains using the ACME protocol.
package domains

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"shiroxy/pkg/models"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/mholt/acmez/acme"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Storage is responsible for handling domain data storage, ACME certificates, and DNS challenge tokens.
type Storage struct {
	WebhookSecret        string                     // Secret for webhook verification.
	ACME_SERVER_URL      string                     // URL for the ACME server.
	INSECURE_SKIP_VERIFY bool                       // Flag to skip SSL verification; used for testing only.
	storage              *models.Storage            // Storage configuration (e.g., memory, Redis).
	redisClient          *redis.Client              // Redis client for interacting with Redis storage.
	DnsChallengeToken    map[string]string          // Map to store DNS challenge tokens for domains.
	DomainMetadata       map[string]*DomainMetadata // Metadata for each registered domain.
}

// InitializeStorage initializes the storage system for managing domain metadata.
// Parameters:
//   - storage: *models.Storage, configuration for storage (memory or Redis).
//   - acmeServerUrl: string, URL for the ACME server.
//   - insecureSkipVerify: string, determines whether to skip SSL verification ("yes" or "no").
//   - wg: *sync.WaitGroup, used for synchronization.
//
// Returns:
//   - *Storage: a pointer to the initialized Storage instance.
//   - error: error if the storage could not be initialized.
func InitializeStorage(storage *models.Storage, acmeServerUrl string, insecureSkipVerify string, wg *sync.WaitGroup) (*Storage, error) {
	if acmeServerUrl == "" {
		acmeServerUrl = "https://127.0.0.1:14000/dir"
	}

	storageSystem := Storage{
		storage:              storage,
		DnsChallengeToken:    make(map[string]string),
		ACME_SERVER_URL:      acmeServerUrl,
		INSECURE_SKIP_VERIFY: true, // Defaults to true unless explicitly set to "no".
	}

	// Set INSECURE_SKIP_VERIFY based on input.
	if insecureSkipVerify == "no" {
		storageSystem.INSECURE_SKIP_VERIFY = false
	}

	// Initialize storage based on the specified storage location (memory or Redis).
	if storage.Location == "redis" {
		redisClient, err := storageSystem.connectRedis()
		if err != nil {
			return nil, err
		}
		storageSystem.redisClient = redisClient
		storageSystem.DomainMetadata = make(map[string]*DomainMetadata)
	} else if storage.Location == "memory" {
		memoryStorage, err := storageSystem.initiazeMemoryStorage()
		if err != nil {
			return nil, err
		}
		storageSystem.DomainMetadata = memoryStorage
	}
	return &storageSystem, nil
}

// RegisterDomain registers a new domain with metadata and generates an ACME account key.
// Parameters:
//   - domainName: string, the domain to register.
//   - user_email: string, the user's email associated with the domain.
//   - metadata: map[string]string, additional metadata for the domain.
//
// Returns:
//   - string: the DNS challenge key.
//   - error: error if registration fails.
func (s *Storage) RegisterDomain(domainName, user_email string, metadata map[string]string) (string, error) {
	if len(domainName) == 0 {
		return "", errors.New("domainName should not be empty")
	}

	// Generate ACME account keys for the domain.
	domainMetadata, err := s.generateAcmeAccountKeys(domainName, user_email, metadata)
	if err != nil {
		return "", err
	}

	// Store the domain metadata in the appropriate storage (memory or Redis).
	if s.storage.Location == "memory" {
		s.DomainMetadata[domainName] = domainMetadata
	} else if s.storage.Location == "redis" {
		marshaledBody, err := proto.Marshal(domainMetadata)
		if err != nil {
			return "", err
		}

		ctx := context.Background()
		result := s.redisClient.Set(ctx, domainName, marshaledBody, 0)
		if result.Err() != nil {
			return "", result.Err()
		}
	}

	// Generate the certificate for the domain.
	err = s.generateCertificate(domainMetadata)
	if err != nil {
		return "", err
	}

	return domainMetadata.DnsChallengeKey, nil
}

// UpdateDomain updates the metadata for an existing domain.
// Parameters:
//   - domainName: string, the domain to update.
//   - updateBody: *DomainMetadata, the new metadata to update.
//
// Returns:
//   - error: error if the domain cannot be updated.
func (s *Storage) UpdateDomain(domainName string, updateBody *DomainMetadata) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

	// Update the metadata in the appropriate storage (memory or Redis).
	if s.storage.Location == "memory" {
		oldData := s.DomainMetadata[domainName]
		if oldData == nil {
			return errors.New("no data found for domainName")
		} else {
			s.DomainMetadata[domainName] = updateBody
		}
	} else if s.storage.Location == "redis" {
		ctx := context.Background()
		result := s.redisClient.Get(ctx, domainName)
		if result.Err() != nil {
			return result.Err()
		}

		oldData, err := result.Result()
		if err != nil {
			return err
		}

		if oldData == "" {
			return errors.New("no data found for domainName")
		}
		marshaledUpdateBody, err := proto.Marshal(updateBody)
		if err != nil {
			return err
		}
		updateResult := s.redisClient.Set(ctx, domainName, marshaledUpdateBody, 0)
		if updateResult.Err() != nil {
			return updateResult.Err()
		}
	}
	return nil
}

// RemoveDomain removes a domain's metadata from the storage.
// Parameters:
//   - domainName: string, the domain to remove.
//
// Returns:
//   - error: error if the domain cannot be removed.
func (s *Storage) RemoveDomain(domainName string) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

	// Remove domain metadata from memory or Redis.
	if s.storage.Location == "memory" {
		oldData := s.DomainMetadata[domainName]
		if oldData == nil {
			return errors.New("no data found for domainName")
		} else {
			delete(s.DomainMetadata, domainName)
		}
	} else if s.storage.Location == "redis" {
		ctx := context.Background()
		result := s.redisClient.Get(ctx, domainName)
		if result.Err() != nil {
			return result.Err()
		}

		oldData, err := result.Result()
		if err != nil {
			return err
		}

		if oldData == "" {
			return errors.New("no data found for domainName")
		}
		deleteResult := s.redisClient.Del(ctx, domainName)
		if deleteResult.Err() != nil {
			return deleteResult.Err()
		}
	}
	return nil
}

// ForceSSL is a placeholder function to enforce SSL for a domain.
// Parameters:
//   - domainName: string, the domain to enforce SSL on.
//
// Returns:
//   - error: currently, this function returns nil as it's a placeholder.
func (s *Storage) ForceSSL(domainName string) error {
	return nil
}

// connectRedis establishes a connection to the Redis database using the provided configuration.
// Returns:
//   - *redis.Client: the initialized Redis client.
//   - error: error if the connection fails.
func (s *Storage) connectRedis() (*redis.Client, error) {
	var rdb redis.Client

	if s.storage.RedisConnectionString != "" {
		opt, err := redis.ParseURL(s.storage.RedisConnectionString)
		if err != nil {
			panic(err)
		}
		rdb = *redis.NewClient(opt)
	} else {
		rdb = *redis.NewClient(&redis.Options{
			Addr:     s.storage.RedisHost + ":" + s.storage.RedisHost,
			Password: s.storage.RedisPassword,
			DB:       0,
		})
	}

	var ctx context.Context = context.Background()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	} else {
		return &rdb, nil
	}
}

// initiazeMemoryStorage initializes an in-memory storage map for domain metadata.
// Returns:
//   - map[string]*DomainMetadata: the initialized in-memory storage map.
//   - error: currently, this function returns nil as it always succeeds.
func (s *Storage) initiazeMemoryStorage() (map[string]*DomainMetadata, error) {
	memoryMap := map[string]*DomainMetadata{}
	return memoryMap, nil
}

// generateAcmeAccountKeys generates an ACME account private key for a domain.
// Parameters:
//   - domainName: string, the domain to generate keys for.
//   - email: string, the user's email.
//   - metadata: map[string]string, additional metadata for the domain.
//
// Returns:
//   - *DomainMetadata: the generated domain metadata containing the keys.
//   - error: error if key generation fails.
func (s *Storage) generateAcmeAccountKeys(domainName, email string, metadata map[string]string) (*DomainMetadata, error) {
	var domainMetadata *DomainMetadata

	domainMetadata = s.DomainMetadata[domainName]
	if domainMetadata == nil {
		accountPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating account key: %v", err)
		}

		privateKeyBytes, err := x509.MarshalECPrivateKey(accountPrivateKey)
		if err != nil {
			return nil, err
		}

		domainMetadata = &DomainMetadata{
			Status:                "inactive",
			Domain:                domainName,
			Email:                 email,
			Metadata:              metadata,
			AcmeAccountPrivateKey: privateKeyBytes,
		}
	}

	return domainMetadata, nil
}

// generateCertificate generates an SSL certificate for the domain using the ACME protocol.
// Parameters:
//   - domainMetadata: *DomainMetadata, the metadata of the domain to generate a certificate for.
//
// Returns:
//   - error: error if certificate generation fails.
func (s *Storage) generateCertificate(domainMetadata *DomainMetadata) error {
	accountPrivateKey, err := x509.ParseECPrivateKey(domainMetadata.AcmeAccountPrivateKey)
	if err != nil {
		return err
	}
	savedAccount := acme.Account{
		Contact:              []string{fmt.Sprintf("mailto:%s", domainMetadata.Email)},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

	ctx := context.Background()

	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	tlcClientConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if os.Getenv("SHIROXY_ENVIRONMENT") == "stage" || os.Getenv("SHIROXY_ENVIRONMENT") == "prod" {
		tlcClientConfig.InsecureSkipVerify = false
	}

	// Create a low-level ACME client.
	client := &acme.Client{
		Directory: s.ACME_SERVER_URL,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlcClientConfig,
			},
		},
		Logger: logger,
	}

	var account acme.Account

	// Retrieve or create an ACME account.
	account, err = client.GetAccount(ctx, savedAccount)
	if err != nil {
		account, err = client.NewAccount(ctx, savedAccount)
		if err != nil {
			return fmt.Errorf("new account: %v", err)
		}
	}

	// Create a new ACME order for the certificate.
	var ids []acme.Identifier
	ids = append(ids, acme.Identifier{Type: "dns", Value: domainMetadata.Domain})
	order := acme.Order{Identifiers: ids}
	order, err = client.NewOrder(ctx, account, order)
	if err != nil {
		return fmt.Errorf("creating new order: %v", err)
	}

	// Solve the challenges for the domain to authorize certificate issuance.
	for _, authzURL := range order.Authorizations {
		authz, err := client.GetAuthorization(ctx, account, authzURL)
		if err != nil {
			return fmt.Errorf("getting authorization %q: %v", authzURL, err)
		}

		var preferredChallenge acme.Challenge
		for _, challenges := range authz.Challenges {
			if challenges.Type == "http-01" {
				preferredChallenge = challenges
				break
			}
		}

		s.DnsChallengeToken[preferredChallenge.Token] = domainMetadata.Domain
		domainMetadata.DnsChallengeKey = preferredChallenge.KeyAuthorization

		// Initiate the challenge to start solving it.
		preferredChallenge, err = client.InitiateChallenge(ctx, account, preferredChallenge)
		if err != nil {
			return fmt.Errorf("initiating challenge %q: %v", preferredChallenge.URL, err)
		}

		// Poll the authorization until it is valid.
		authz, err = client.PollAuthorization(ctx, account, authz)
		if err != nil {
			return fmt.Errorf("solving challenge: %v", err)
		}
	}

	// Generate a private key for the certificate.
	certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating certificate key: %v", err)
	}

	// Create a certificate signing request (CSR).
	csrTemplate := &x509.CertificateRequest{DNSNames: []string{domainMetadata.Domain}}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, certPrivateKey)
	if err != nil {
		return fmt.Errorf("generating CSR: %v", err)
	}

	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return fmt.Errorf("parsing generated CSR: %v", err)
	}

	// Finalize the ACME order with the CSR to obtain the certificate.
	order, err = client.FinalizeOrder(ctx, account, order, csr.Raw)
	if err != nil {
		return fmt.Errorf("finalizing order: %v", err)
	}

	// Download the certificate chain from the ACME server.
	certChains, err := client.GetCertificateChain(ctx, account, order.Certificate)
	if err != nil {
		return fmt.Errorf("downloading certs: %v", err)
	}

	// Store the certificate and private key in the domain metadata.
	var fullChain []byte
	for _, cert := range certChains {
		fullChain = append(fullChain, cert.ChainPEM...)
	}

	// Marshal the private key to DER format.
	certPrivateKeyBytes, err := x509.MarshalECPrivateKey(certPrivateKey)
	if err != nil {
		return fmt.Errorf("marshaling private key: %v", err)
	}

	// Encode the private key to PEM format.
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: certPrivateKeyBytes,
	})

	// Update the domain metadata with the certificate and key.
	domainMetadata.CertPemBlock = fullChain
	domainMetadata.KeyPemBlock = privateKeyPEM
	domainMetadata.Metadata = make(map[string]string)
	domainMetadata.Metadata["cert_url"] = certChains[0].URL
	domainMetadata.Metadata["cert_ca"] = certChains[0].CA
	domainMetadata.Status = "active"

	fmt.Printf("Certificate Generated Successfully For Domain : %s", domainMetadata.Domain)

	return nil
}
