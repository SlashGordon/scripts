package security_test

import (
	"strings"
	"testing"
)

func TestTorExitAddressParsing(t *testing.T) {
	torData := `# This is a comment
ExitAddress 192.168.1.1 2023-01-01 12:00:00
ExitAddress 10.0.0.1 2023-01-01 12:01:00
# Another comment
ExitAddress 203.0.113.1 2023-01-01 12:02:00
Invalid line without ExitAddress
ExitAddress 198.51.100.1 2023-01-01 12:03:00`

	expected := []string{
		"iptables -A BLOCKLIST -s 192.168.1.1 -j DROP",
		"iptables -A BLOCKLIST -s 10.0.0.1 -j DROP",
		"iptables -A BLOCKLIST -s 203.0.113.1 -j DROP",
		"iptables -A BLOCKLIST -s 198.51.100.1 -j DROP",
	}

	rules := parseTorExitAddresses(torData, "BLOCKLIST")

	if len(rules) != len(expected) {
		t.Errorf("Expected %d rules, got %d", len(expected), len(rules))
		return
	}

	for i, rule := range rules {
		if rule != expected[i] {
			t.Errorf("Rule %d: expected '%s', got '%s'", i, expected[i], rule)
		}
	}
}

func parseTorExitAddresses(torData, chain string) []string {
	var rules []string

	for _, line := range strings.Split(torData, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "ExitAddress ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := parts[1]
				rule := "iptables -A " + chain + " -s " + ip + " -j DROP"
				rules = append(rules, rule)
			}
		}
	}

	return rules
}
