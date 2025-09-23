package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ddnsCmd = &cobra.Command{
	Use:   "ddns",
	Short: "Cloudflare DDNS updater",
	Long:  "Update Cloudflare DNS records with current public IP",
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update DNS records",
	Run: func(cmd *cobra.Command, args []string) {
		config := getDDNSConfig()
		
		if config.APIToken == "" || config.ZoneID == "" || config.RecordName == "" {
			fmt.Println("Error: CF_API_TOKEN, CF_ZONE_ID, and CF_RECORD_NAME environment variables are required")
			os.Exit(1)
		}
		
		currentIP := getCurrentIP("https://api.ipify.org")
		currentIPv6 := getCurrentIP("https://api6.ipify.org")
		
		if currentIP == "" && currentIPv6 == "" {
			logError("could not get current public IPs", config.LogFile)
			os.Exit(1)
		}
		
		cachedIP4, cachedIP6 := readCache(config.CacheFile)
		
		// Update IPv4
		if currentIP != "" && currentIP != cachedIP4 {
			if updateRecord(config, "A", currentIP) {
				cachedIP4 = currentIP
				logInfo(fmt.Sprintf("Updated A %s → %s", config.RecordName, currentIP), config.LogFile)
			}
		} else {
			logInfo(fmt.Sprintf("IPv4 unchanged (%s)", currentIP), config.LogFile)
		}
		
		// Update IPv6
		if currentIPv6 != "" && currentIPv6 != cachedIP6 {
			if updateRecord(config, "AAAA", currentIPv6) {
				cachedIP6 = currentIPv6
				logInfo(fmt.Sprintf("Updated AAAA %s → %s", config.RecordName, currentIPv6), config.LogFile)
			}
		} else {
			logInfo(fmt.Sprintf("IPv6 unchanged (%s)", currentIPv6), config.LogFile)
		}
		
		writeCache(config.CacheFile, cachedIP4, cachedIP6)
	},
}

type DDNSConfig struct {
	APIToken   string
	ZoneID     string
	RecordName string
	LogFile    string
	CacheFile  string
}

func getDDNSConfig() DDNSConfig {
	return DDNSConfig{
		APIToken:   getEnv("CF_API_TOKEN", ""),
		ZoneID:     getEnv("CF_ZONE_ID", ""),
		RecordName: getEnv("CF_RECORD_NAME", ""),
		LogFile:    getEnv("DDNS_LOG_FILE", "./ddns.log"),
		CacheFile:  getEnv("DDNS_CACHE_FILE", "./.ddns.cache"),
	}
}

func getCurrentIP(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(body))
}

func updateRecord(config DDNSConfig, recordType, ip string) bool {
	recordID := getRecordID(config, recordType)
	if recordID == "" {
		logError(fmt.Sprintf("%s record %s not found", recordType, config.RecordName), config.LogFile)
		return false
	}
	
	data := map[string]interface{}{
		"type":    recordType,
		"name":    config.RecordName,
		"content": ip,
		"proxied": true,
	}
	
	jsonData, _ := json.Marshal(data)
	
	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.ZoneID, recordID), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+config.APIToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logError(fmt.Sprintf("%s update failed: %v", recordType, err), config.LogFile)
		return false
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if success, ok := result["success"].(bool); ok && success {
		return true
	}
	
	logError(fmt.Sprintf("%s update failed", recordType), config.LogFile)
	return false
}

func getRecordID(config DDNSConfig, recordType string) string {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=%s", config.ZoneID, config.RecordName, recordType)
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+config.APIToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if resultArray, ok := result["result"].([]interface{}); ok && len(resultArray) > 0 {
		if record, ok := resultArray[0].(map[string]interface{}); ok {
			if id, ok := record["id"].(string); ok {
				return id
			}
		}
	}
	
	return ""
}

func readCache(cacheFile string) (string, string) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", ""
	}
	
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) >= 2 {
		return lines[0], lines[1]
	}
	if len(lines) == 1 {
		return lines[0], ""
	}
	
	return "", ""
}

func writeCache(cacheFile, ip4, ip6 string) {
	content := ip4 + "\n" + ip6
	os.WriteFile(cacheFile, []byte(content), 0644)
}

func logInfo(msg, logFile string) {
	logMessage("INFO", msg, logFile)
}

func logError(msg, logFile string) {
	logMessage("ERROR", msg, logFile)
}

func logMessage(level, msg, logFile string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("%s %s: %s\n", timestamp, level, msg)
	
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print(logLine)
		return
	}
	defer f.Close()
	
	f.WriteString(logLine)
}

func init() {
	ddnsCmd.AddCommand(updateCmd)
}