package security_test

import (
	"os"
	"testing"

	"github.com/SlashGordon/nas-manager/cmd/security"
)

func TestGetBlocklists(t *testing.T) {
	// Test default blocklists
	lists := security.GetBlocklists()
	if len(lists) != 3 {
		t.Errorf("Expected 3 default blocklists, got %d", len(lists))
	}

	expected := []string{"firehol_level1", "spamhaus_drop", "dshield"}
	for i, list := range lists {
		if list.Name != expected[i] {
			t.Errorf("Expected blocklist name '%s', got '%s'", expected[i], list.Name)
		}
	}
}

func TestGetBlocklistsWithCustom(t *testing.T) {
	oldCustom := os.Getenv("SECURITY_CUSTOM_LISTS")
	defer t.Setenv("SECURITY_CUSTOM_LISTS", oldCustom)

	t.Setenv("SECURITY_CUSTOM_LISTS", "custom1=http://example.com/list1,custom2=http://example.com/list2")

	lists := security.GetBlocklists()
	if len(lists) != 2 {
		t.Errorf("Expected 2 custom blocklists (defaults disabled), got %d", len(lists))
	}

	if lists[0].Name != "custom1" || lists[1].Name != "custom2" {
		t.Errorf("Custom blocklists not parsed correctly")
	}
}

func TestGetBlocklistsInvalidCustom(t *testing.T) {
	oldCustom := os.Getenv("SECURITY_CUSTOM_LISTS")
	defer t.Setenv("SECURITY_CUSTOM_LISTS", oldCustom)

	t.Setenv("SECURITY_CUSTOM_LISTS", "invalid_format")

	lists := security.GetBlocklists()
	if len(lists) != 0 {
		t.Errorf("Expected 0 blocklists with invalid custom format, got %d", len(lists))
	}
}
