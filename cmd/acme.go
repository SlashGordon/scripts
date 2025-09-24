package cmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spf13/cobra"
)

var acmeCmd = &cobra.Command{
	Use:   "acme",
	Short: "ACME certificate management",
	Long:  "Issue and renew Let's Encrypt certificates via Lego and Cloudflare DNS-01 challenge",
}

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issue/renew certificate",
	Run: func(cmd *cobra.Command, args []string) {
		config := getAcmeConfig()
		
		if config.CFToken == "" || config.Domain == "" || config.Email == "" {
			fmt.Println("Error: CF_API_TOKEN, ACME_DOMAIN, and ACME_EMAIL environment variables are required")
			os.Exit(1)
		}
		
		fmt.Printf("Issuing certificate for domain: %s\n", config.Domain)
		
		if err := issueCertificate(config); err != nil {
			fmt.Printf("Certificate issue failed: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Certificate for %s issued successfully.\n", config.Domain)
	},
}

type AcmeConfig struct {
	Domain   string
	CertPath string
	Email    string
	CFToken  string
}

func getAcmeConfig() AcmeConfig {
	return AcmeConfig{
		Domain:   getEnv("ACME_DOMAIN", ""),
		CertPath: getEnv("ACME_CERT_PATH", "./cert"),
		Email:    getEnv("ACME_EMAIL", ""),
		CFToken:  getEnv("CF_API_TOKEN", ""),
	}
}

func issueCertificate(config AcmeConfig) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	
	user := &User{
		Email: config.Email,
		key:   privateKey,
	}
	
	legoConfig := lego.NewConfig(user)
	legoConfig.Certificate.KeyType = certcrypto.EC256
	
	client, err := lego.NewClient(legoConfig)
	if err != nil {
		return err
	}
	
	os.Setenv("CLOUDFLARE_DNS_API_TOKEN", config.CFToken)
	provider, err := cloudflare.NewDNSProvider()
	if err != nil {
		return err
	}
	
	client.Challenge.SetDNS01Provider(provider)
	
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	user.Registration = reg
	
	request := certificate.ObtainRequest{
		Domains: []string{config.Domain},
		Bundle:  true,
	}
	
	certs, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	
	certPath := config.CertPath
	usingFallback := false
	
	// Try to create directory and test write permissions
	if err := os.MkdirAll(certPath, 0755); err != nil {
		usingFallback = true
	} else {
		// Test write permissions by creating a temporary file
		testFile := filepath.Join(certPath, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			usingFallback = true
		} else {
			os.Remove(testFile)
		}
	}
	
	if usingFallback {
		dateStr := time.Now().Format("2006-01-02")
		domainSafe := strings.ReplaceAll(config.Domain, ".", "_")
		certPath = fmt.Sprintf("./certs/%s-%s", domainSafe, dateStr)
		fmt.Printf("Permission denied for %s, using fallback directory: %s\n", config.CertPath, certPath)
		if err := os.MkdirAll(certPath, 0755); err != nil {
			return fmt.Errorf("failed to create fallback directory: %v", err)
		}
	}
	
	// Write certificates
	files := map[string][]byte{
		config.Domain + ".key": certs.PrivateKey,
		config.Domain + ".cer": certs.Certificate,
		"fullchain.cer":        certs.Certificate,
		"ca.cer":              certs.IssuerCertificate,
		"privkey.pem":         certs.PrivateKey,
		"fullchain.pem":       certs.Certificate,
	}
	
	for filename, content := range files {
		perm := os.FileMode(0644)
		if filepath.Ext(filename) == ".key" || filename == "privkey.pem" {
			perm = 0600
		}
		if err := os.WriteFile(filepath.Join(certPath, filename), content, perm); err != nil {
			return fmt.Errorf("failed to write %s: %v", filename, err)
		}
	}
	
	if usingFallback {
		fmt.Printf("Certificates saved to fallback directory: %s\n", certPath)
		fmt.Printf("Please manually copy certificates to: %s\n", config.CertPath)
		fmt.Printf("Run: sudo cp %s/* %s/\n", certPath, config.CertPath)
	} else {
		exec.Command("/usr/syno/sbin/synoservicectl", "--reload", "nginx").Run()
	}
	
	return nil
}

type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func init() {
	acmeCmd.AddCommand(issueCmd)
}