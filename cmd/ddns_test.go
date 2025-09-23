package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetCurrentIP(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("192.168.1.100"))
	}))
	defer server.Close()
	
	ip := getCurrentIP(server.URL)
	if ip != "192.168.1.100" {
		t.Errorf("Expected '192.168.1.100', got '%s'", ip)
	}
}

func TestGetCurrentIPError(t *testing.T) {
	// Test with invalid URL
	ip := getCurrentIP("http://invalid-url-that-does-not-exist")
	if ip != "" {
		t.Errorf("Expected empty string for invalid URL, got '%s'", ip)
	}
}

func TestReadWriteCache(t *testing.T) {
	cacheFile := "/tmp/test_ddns_cache"
	
	// Test writing cache
	writeCache(cacheFile, "192.168.1.1", "2001:db8::1")
	
	// Test reading cache
	ip4, ip6 := readCache(cacheFile)
	
	if ip4 != "192.168.1.1" {
		t.Errorf("Expected '192.168.1.1', got '%s'", ip4)
	}
	if ip6 != "2001:db8::1" {
		t.Errorf("Expected '2001:db8::1', got '%s'", ip6)
	}
	
	// Cleanup
	os.Remove(cacheFile)
}

func TestReadCacheEmpty(t *testing.T) {
	// Test reading non-existent cache
	ip4, ip6 := readCache("/tmp/non_existent_cache")
	if ip4 != "" || ip6 != "" {
		t.Errorf("Expected empty strings for non-existent cache, got '%s', '%s'", ip4, ip6)
	}
}

func TestReadCacheSingleLine(t *testing.T) {
	cacheFile := "/tmp/test_single_cache"
	
	// Write single line cache
	os.WriteFile(cacheFile, []byte("192.168.1.1"), 0644)
	
	ip4, ip6 := readCache(cacheFile)
	if ip4 != "192.168.1.1" || ip6 != "" {
		t.Errorf("Expected '192.168.1.1' and empty string, got '%s', '%s'", ip4, ip6)
	}
	
	os.Remove(cacheFile)
}