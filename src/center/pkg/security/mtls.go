// Package security provides enterprise security features
package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"
)

// MTLSConfig holds mTLS configuration
type MTLSConfig struct {
	Enabled            bool   `json:"enabled"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	CAFile             string `json:"ca_file"`
	ServerName         string `json:"server_name"`
	ClientAuth         bool   `json:"client_auth"` // Require client certificates
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

// MTLSManager manages mutual TLS
type MTLSManager struct {
	config MTLSConfig

	// Certificate pool
	certPool *x509.CertPool

	// Server certificate
	serverCert tls.Certificate

	// Client certificates
	clientCerts sync.Map // map[string]*x509.Certificate

	// Certificate revocation
	revokedCerts sync.Map // map[string]bool

	mu sync.RWMutex
}

// NewMTLSManager creates a new mTLS manager
func NewMTLSManager(config MTLSConfig) (*MTLSManager, error) {
	manager := &MTLSManager{
		config: config,
	}

	if !config.Enabled {
		return manager, nil
	}

	// Load CA certificate
	if config.CAFile != "" {
		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA cert: %v", err)
		}

		manager.certPool = x509.NewCertPool()
		if !manager.certPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to parse CA cert")
		}
	}

	// Load server certificate
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load server cert: %v", err)
		}
		manager.serverCert = cert
	}

	return manager, nil
}

// GetTLSConfig returns TLS configuration for server
func (m *MTLSManager) GetTLSConfig() *tls.Config {
	if !m.config.Enabled {
		return nil
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{m.serverCert},
		ServerName:   m.config.ServerName,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	if m.config.ClientAuth {
		tlsConfig.ClientCAs = m.certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if m.config.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig
}

// GetClientTLSConfig returns TLS configuration for client
func (m *MTLSManager) GetClientTLSConfig() *tls.Config {
	if !m.config.Enabled {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{m.serverCert},
		RootCAs:            m.certPool,
		ServerName:         m.config.ServerName,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: m.config.InsecureSkipVerify,
	}

	return tlsConfig
}

// AddClientCert adds a client certificate
func (m *MTLSManager) AddClientCert(clientID string, certPEM []byte) error {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return errors.New("failed to parse certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	m.clientCerts.Store(clientID, cert)
	return nil
}

// RemoveClientCert removes a client certificate
func (m *MTLSManager) RemoveClientCert(clientID string) {
	m.clientCerts.Delete(clientID)
}

// RevokeCertificate revokes a certificate
func (m *MTLSManager) RevokeCertificate(serialNumber string) {
	m.revokedCerts.Store(serialNumber, true)
}

// IsRevoked checks if a certificate is revoked
func (m *MTLSManager) IsRevoked(serialNumber string) bool {
	_, ok := m.revokedCerts.Load(serialNumber)
	return ok
}

// VerifyClient verifies client certificate
func (m *MTLSManager) VerifyClient(cert *x509.Certificate) error {
	// Check if revoked
	if m.IsRevoked(cert.SerialNumber.String()) {
		return errors.New("certificate has been revoked")
	}

	// Check expiry
	if time.Now().After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}

	// Check not before
	if time.Now().Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}

	return nil
}

// CertificateAuthority manages certificate generation
type CertificateAuthority struct {
	caCert   *x509.Certificate
	caKey    *rsa.PrivateKey
	certPool *x509.CertPool

	certPath string
	keyPath  string

	mu sync.RWMutex
}

// NewCertificateAuthority creates a new CA
func NewCertificateAuthority(certPath, keyPath string) (*CertificateAuthority, error) {
	ca := &CertificateAuthority{
		certPath: certPath,
		keyPath:  keyPath,
	}

	// Try to load existing CA
	if _, err := os.Stat(certPath); err == nil {
		if err := ca.load(); err != nil {
			return nil, err
		}
	} else {
		// Generate new CA
		if err := ca.generate(); err != nil {
			return nil, err
		}
	}

	return ca, nil
}

// generate generates a new CA certificate
func (ca *CertificateAuthority) generate() error {
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"OFA"},
			CommonName:   "OFA Certificate Authority",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Generate key
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Self-sign
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	ca.caCert, _ = x509.ParseCertificate(certDER)
	ca.caKey = key

	// Save to disk
	if err := ca.save(); err != nil {
		return err
	}

	return nil
}

// load loads CA from disk
func (ca *CertificateAuthority) load() error {
	certPEM, err := os.ReadFile(ca.certPath)
	if err != nil {
		return err
	}

	keyPEM, err := os.ReadFile(ca.keyPath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(certPEM)
	ca.caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	keyBlock, _ := pem.Decode(keyPEM)
	ca.caKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return err
	}

	return nil
}

// save saves CA to disk
func (ca *CertificateAuthority) save() error {
	// Save certificate
	certFile, err := os.Create(ca.certPath)
	if err != nil {
		return err
	}
	defer certFile.Close()

	pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.caCert.Raw,
	})

	// Save key
	keyFile, err := os.Create(ca.keyPath)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(ca.caKey),
	})

	return nil
}

// IssueCertificate issues a new certificate
func (ca *CertificateAuthority) IssueCertificate(commonName string, days int, isServer bool) ([]byte, []byte, error) {
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"OFA"},
			CommonName:   commonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Duration(days) * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{},
	}

	if isServer {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		template.DNSNames = []string{commonName}
	} else {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	// Generate key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Sign with CA
	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.caCert, &key.PublicKey, ca.caKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return certPEM, keyPEM, nil
}