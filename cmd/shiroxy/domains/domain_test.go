package domains_test

import (
	"context"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/pkg/models"
	"sync"
	"testing"
)

func TestInitializeStorage_Memory(t *testing.T) {
	storage := &models.Storage{
		Location: "memory",
	}

	var wg sync.WaitGroup
	store, err := domains.InitializeStorage(storage, "", "yes", &wg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if store == nil {
		t.Fatalf("expected storage to be initialized")
	}

	if store.ACME_SERVER_URL != "https://127.0.0.1:14000/dir" {
		t.Errorf("unexpected ACME_SERVER_URL: got %s", store.ACME_SERVER_URL)
	}

	if len(store.DomainMetadata) != 0 {
		t.Errorf("expected DomainMetadata to be empty, got %v", store.DomainMetadata)
	}
}

func TestInitializeStorage_Redis(t *testing.T) {
	t.Skip("Skipping Redis test - requires running Redis instance")
	storage := &models.Storage{
		Location:              "redis",
		RedisConnectionString: "redis://localhost:6379/0",
	}

	var wg sync.WaitGroup
	store, err := domains.InitializeStorage(storage, "https://acme-staging-v02.api.letsencrypt.org/directory", "no", &wg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if store == nil {
		t.Fatalf("expected storage to be initialized")
	}

	if store.INSECURE_SKIP_VERIFY {
		t.Errorf("expected INSECURE_SKIP_VERIFY to be false, got true")
	}

	if store.RedisClient == nil {
		t.Errorf("expected Redis client to be initialized")
	}
}

func TestRegisterDomain_EmptyDomain(t *testing.T) {
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location: "memory",
		},
		DomainMetadata: make(map[string]*domains.DomainMetadata),
	}

	_, err := storage.RegisterDomain("", "user@example.com", nil)
	if err == nil || err.Error() != "domainName should not be empty" {
		t.Errorf("expected error 'domainName should not be empty', got %v", err)
	}
}

func TestRegisterDomain_Memory(t *testing.T) {
	t.Skip("Skipping RegisterDomain test - requires ACME directory configuration")
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location: "memory",
		},
		DomainMetadata: make(map[string]*domains.DomainMetadata),
	}

	dnsChallengeKey, err := storage.RegisterDomain("example.com", "user@example.com", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(dnsChallengeKey) == 0 {
		t.Errorf("expected DNS challenge key to be generated")
	}

	if storage.DomainMetadata["example.com"] == nil {
		t.Errorf("expected domain metadata to be stored in memory")
	}
}

func TestRegisterDomain_Redis(t *testing.T) {
	t.Skip("Skipping Redis test - requires running Redis instance")
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location:              "redis",
			RedisConnectionString: "redis://localhost:6379/0",
		},
		DomainMetadata: make(map[string]*domains.DomainMetadata),
	}

	redisClient, err := storage.ConnectRedis()
	if err != nil {
		t.Fatalf("unable to connect to redis: %v", err)
	}
	storage.RedisClient = redisClient

	dnsChallengeKey, err := storage.RegisterDomain("example.com", "user@example.com", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(dnsChallengeKey) == 0 {
		t.Errorf("expected DNS challenge key to be generated")
	}

	ctx := context.Background()
	result := redisClient.Get(ctx, "example.com")
	if result.Err() != nil {
		t.Errorf("expected domain metadata to be stored in Redis")
	}
}

func TestUpdateDomain_EmptyDomain(t *testing.T) {
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location: "memory",
		},
		DomainMetadata: make(map[string]*domains.DomainMetadata),
	}

	err := storage.UpdateDomain("", &domains.DomainMetadata{})
	if err == nil || err.Error() != "domainName should not be empty" {
		t.Errorf("expected error 'domainName should not be empty', got %v", err)
	}
}

func TestUpdateDomain_Memory(t *testing.T) {
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location: "memory",
		},
		DomainMetadata: map[string]*domains.DomainMetadata{
			"example.com": {Domain: "example.com", Status: "inactive"},
		},
	}

	updateBody := &domains.DomainMetadata{Domain: "example.com", Status: "active"}
	err := storage.UpdateDomain("example.com", updateBody)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storage.DomainMetadata["example.com"].Status != "active" {
		t.Errorf("expected domain status to be updated to 'active'")
	}
}

func TestRemoveDomain_EmptyDomain(t *testing.T) {
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location: "memory",
		},
		DomainMetadata: make(map[string]*domains.DomainMetadata),
	}

	err := storage.RemoveDomain("")
	if err == nil || err.Error() != "domainName should not be empty" {
		t.Errorf("expected error 'domainName should not be empty', got %v", err)
	}
}

func TestRemoveDomain_Memory(t *testing.T) {
	storage := &domains.Storage{
		Storage: &models.Storage{
			Location: "memory",
		},
		DomainMetadata: map[string]*domains.DomainMetadata{
			"example.com": {Domain: "example.com", Status: "inactive"},
		},
	}

	err := storage.RemoveDomain("example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storage.DomainMetadata["example.com"] != nil {
		t.Errorf("expected domain metadata to be removed from memory")
	}
}
