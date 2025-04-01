package certer

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"diploma/models"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

type Interface interface {
	GenerateCertSignedByCA(domain string) (string, string, error)
}

type CertInfo struct {
	ValidFor     time.Duration `yaml:"valid_for"`
	CaCert       string        `yaml:"ca_cert"`
	CaKey        string        `yaml:"ca_key"`
	Organization string        `yaml:"organization"`
}

type Config struct {
	Certificates map[models.Domain]CertInfo `yaml:"certificates"`
}

type Certer struct {
	conf Config
}

func New(conf Config) *Certer {
	return &Certer{
		conf: conf,
	}
}

func (c *Certer) GenerateCertSignedByCA(domain string) (string, string, error) {
	certInfo, ok := c.conf.Certificates[models.Domain(domain)]
	if !ok {
		return "", "", fmt.Errorf("no certificate configuration found for domain %s", domain)
	}

	caBlock, _ := pem.Decode([]byte(certInfo.CaCert))
	if caBlock == nil {
		return "", "", fmt.Errorf("failed to decode CA certificate (PEM)")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	caKeyBlock, _ := pem.Decode([]byte(certInfo.CaKey))
	if caKeyBlock == nil {
		return "", "", fmt.Errorf("failed to decode CA private key (PEM)")
	}
	caKey, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA private key: %w", err)
	}

	newKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new private key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(certInfo.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{certInfo.Organization},
			CommonName:   domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	template.DNSNames = []string{
		domain,
	}
	if ip := net.ParseIP(domain); ip != nil {
		template.IPAddresses = []net.IP{ip}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &newKey.PublicKey, caKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	var certPEM, keyPEM bytes.Buffer
	if err = pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}
	if err = pem.Encode(&keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(newKey)}); err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	log.Printf("Certificate for domain %s has been signed and generated", domain)
	return certPEM.String(), keyPEM.String(), nil
}
