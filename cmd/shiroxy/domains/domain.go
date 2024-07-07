package domains

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"shiroxy/pkg/models"

	"github.com/go-redis/redis/v8"
	"github.com/mholt/acmez/acme"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Storage struct {
	storage        *models.Storage
	redisClient    *redis.Client
	DomainMetadata map[string]*DomainMetadata
}

func InitializeStorage(storage *models.Storage) (*Storage, error) {
	storageSystem := Storage{
		storage: storage,
	}
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

func (s *Storage) RegisterDomain(domainName, user_email string, metadata map[string]string) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

	domainMetadata, err := s.generateAcmeAccountKeys(domainName, user_email, metadata)
	if err != nil {
		return err
	}

	if s.storage.Location == "memory" {
		s.DomainMetadata[domainName] = domainMetadata
	} else if s.storage.Location == "redis" {
		marshaledBody, err := proto.Marshal(domainMetadata)
		if err != nil {
			return err
		}

		ctx := context.Background()
		result := s.redisClient.Set(ctx, domainName, marshaledBody, 0)
		if result.Err() != nil {
			return result.Err()
		}
	}

	// TODO: June 30 10:34 AM - Handle Certificate Storage, generation is already handled.
	s.generateCertificate(domainMetadata)

	return nil
}

func (s *Storage) UpdateDomain(domainName string, updateBody *DomainMetadata) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

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

func (s *Storage) RemoveDomain(domainName string) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

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

func (s *Storage) ForceSSL(domainName string) error {
	return nil
}

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

func (s *Storage) initiazeMemoryStorage() (map[string]*DomainMetadata, error) {
	memoryMap := map[string]*DomainMetadata{}
	return memoryMap, nil
}

type DomainCertificateMetadata struct {
	Account  acme.Account
	CSRDer   []byte
	Domains  []string
	Metadata map[string]string
}

func (s *Storage) generateAcmeAccountKeys(domainName, email string, metadata map[string]string) (*DomainMetadata, error) {
	// put your domains here (IDNs must be in ASCII form)
	domains := []string{domainName}

	var csrDER []byte
	var accountPrivateKey *ecdsa.PrivateKey
	var domainMetadata *DomainMetadata

	domainMetadata = s.DomainMetadata[domainName]
	if domainMetadata != nil {
		var err error
		csrDER = domainMetadata.CsrDer
		accountPrivateKey, err = x509.ParseECPrivateKey(domainMetadata.AcmeAccountPrivateKey)
		if err != nil {
			return nil, err
		}
	} else {
		// first you need a private key for your certificate
		certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating certificate key: %v", err)
		}

		// then you need a certificate request; here's a simple one - we need
		// to fill out the template, then create the actual CSR, then parse it
		// Certificate Signing Request (CSR)
		csrTemplate := &x509.CertificateRequest{DNSNames: domains}
		csrDER, err = x509.CreateCertificateRequest(rand.Reader, csrTemplate, certPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("generating CSR: %v", err)
		}

		// before you can get a cert, you'll need an account registered with
		// the ACME CA - it also needs a private key and should obviously be
		// different from any key used for certificates!
		accountPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating account key: %v", err)
		}

		privateKeyBytes, err := x509.MarshalECPrivateKey(accountPrivateKey)
		if err != nil {
			return nil, err
		}
		domainMetadata = &DomainMetadata{
			Domain:                domainName,
			Email:                 email,
			Metadata:              metadata,
			CsrDer:                csrDER,
			AcmeAccountPrivateKey: privateKeyBytes,
		}
	}

	return domainMetadata, nil

	// fmt.Println(strings.TrimSpace(email))
	// var emails []string = []string{}
	// // emails = append(emails, "yshikharfzd10@gmail.com")
	// // if len(email) > 0 {
	// // 	emails = append(emails, email)
	// // }

	// fmt.Println("emails ===========")
	// fmt.Println(emails)
	// fmt.Println("length of emails: ", len(emails))

	// account := acme.Account{
	// 	Contact:              emails,
	// 	TermsOfServiceAgreed: true,
	// 	PrivateKey:           accountPrivateKey,
	// }

	// return &DomainCertificateMetadata{
	// 	Domains: []string{domainName},
	// 	CSRDer:  csrDER,

	// 	Account:  account,
	// 	Metadata: metadata,
	// }, nil
}

func (s *Storage) generateCertificate(domainMetadata *DomainMetadata) error {

	accountPrivateKey, err := x509.ParseECPrivateKey(domainMetadata.AcmeAccountPrivateKey)
	if err != nil {
		return err
	}
	savedAccount := acme.Account{
		Contact:              []string{fmt.Sprintf("mailto:%s", &domainMetadata.Email)},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

	// a context allows us to cancel long-running ops
	ctx := context.Background()

	// Logging is important - replace with your own zap logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	// now we can make our low-level ACME client
	client := &acme.Client{
		Directory: "https://127.0.0.1:14000/dir", // default pebble endpoint
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // REMOVE THIS FOR PRODUCTION USE!
				},
			},
		},
		Logger: logger,
	}

	var account acme.Account

	account, err = client.GetAccount(ctx, savedAccount)
	if err != nil {
		// if the account is new, we need to create it; only do this once!
		// then be sure to securely store the account key and metadata so
		// you can reuse it later!
		account, err = client.NewAccount(ctx, savedAccount)
		if err != nil {
			return fmt.Errorf("new account: %v", err)
		}
	}

	// now we can actually get a cert; first step is to create a new order
	var ids []acme.Identifier
	ids = append(ids, acme.Identifier{Type: "dns", Value: domainMetadata.Domain})
	// for _, domain := range domainCertificateMetadata.Domains {
	// 	ids = append(ids, acme.Identifier{Type: "dns", Value: domain})
	// }

	fmt.Println("2")
	order := acme.Order{Identifiers: ids}
	order, err = client.NewOrder(ctx, account, order)
	if err != nil {
		return fmt.Errorf("creating new order: %v", err)
	}

	fmt.Println("3")

	// each identifier on the order should now be associated with an
	// authorization object; we must make the authorization "valid"
	// by solving any of the challenges offered for it
	for _, authzURL := range order.Authorizations {
		authz, err := client.GetAuthorization(ctx, account, authzURL)
		if err != nil {
			return fmt.Errorf("getting authorization %q: %v", authzURL, err)
		}

		fmt.Println("=========================== authzURL: ", authzURL)

		// pick any available challenge to solve
		challenge := authz.Challenges[0]

		fmt.Println("challenge: ", challenge.Identifier)
		fmt.Println("K ===== ", challenge.KeyAuthorization)
		fmt.Println("T ===== ", challenge.Token)

		// at this point, you must prepare to solve the challenge; how
		// you do this depends on the challenge (see spec for details).
		// usually this involves configuring an HTTP or TLS server, but
		// it might also involve setting a DNS record (which can take
		// time to propagate, depending on the provider!) - this example
		// does NOT do this step for you - it's "bring your own solver."

		// once you are ready to solve the challenge, let the ACME
		// server know it should begin
		challenge, err = client.InitiateChallenge(ctx, account, challenge)
		if err != nil {
			return fmt.Errorf("initiating challenge %q: %v", challenge.URL, err)
		}

		fmt.Println("challenge: ", challenge)

		// now the challenge should be under way; at this point, we can
		// continue initiating all the other challenges so that they are
		// all being solved in parallel (this saves time when you have a
		// large number of SANs on your certificate), but this example is
		// simple, so we will just do one at a time; we wait for the ACME
		// server to tell us the challenge has been solved by polling the
		// authorization status
		authz, err = client.PollAuthorization(ctx, account, authz)
		if err != nil {
			return fmt.Errorf("solving challenge: %v", err)
		}

		// if we got here, then the challenge was solved successfully, hurray!
	}

	fmt.Println("4")

	csr, err := x509.ParseCertificateRequest(domainMetadata.CsrDer)
	if err != nil {
		return fmt.Errorf("parsing generated CSR: %v", err)
	}

	// to request a certificate, we finalize the order; this function
	// will poll the order status for us and return once the cert is
	// ready (or until there is an error)
	order, err = client.FinalizeOrder(ctx, account, order, csr.Raw)
	if err != nil {
		return fmt.Errorf("finalizing order: %v", err)
	}

	// we can now download the certificate; the server should actually
	// provide the whole chain, and it can even offer multiple chains
	// of trust for the same end-entity certificate, so this function
	// returns all of them; you can decide which one to use based on
	// your own requirements
	certChains, err := client.GetCertificateChain(ctx, account, order.Certificate)
	if err != nil {
		return fmt.Errorf("downloading certs: %v", err)
	}

	// all done! store it somewhere safe, along with its key
	for _, cert := range certChains {
		fmt.Printf("Certificate %q:\n%s\n\n", cert.URL, cert.ChainPEM)
	}
	// TODO: June 30 10:34 AM - Handle Certificate Storage, generation is already handled.
	return nil
}
