package domains

import (
	"shiroxy/pkg/models"
	"testing"
)

func TestInitMemoryStorageAndGenerateAccountKey(t *testing.T) {
	st := Storage{Storage: &models.Storage{Location: "memory"}}
	mem, err := st.initiazeMemoryStorage()
	if err != nil {
		t.Fatalf("init memory storage: %v", err)
	}
	if mem == nil {
		t.Fatalf("expected non-nil memory storage map")
	}

	// test generateAcmeAccountKeys (does not call network)
	dm, err := st.generateAcmeAccountKeys("example.com", "a@b.com", map[string]string{"tags": "api"})
	if err != nil {
		t.Fatalf("generateAcmeAccountKeys failed: %v", err)
	}
	if dm.Domain != "example.com" {
		t.Fatalf("unexpected domain: %s", dm.Domain)
	}
	if dm.Status != "inactive" {
		t.Fatalf("expected inactive status on new domain")
	}
}
