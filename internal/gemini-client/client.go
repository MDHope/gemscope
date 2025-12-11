package gemini_client

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GeminiClient struct {
	knownHosts map[string]string
	hostsFile  string
	Timeout    time.Duration
}

func NewGeminiClient() *GeminiClient {
	homeDir, _ := os.UserHomeDir()
	hostsFile := filepath.Join(homeDir, ".gemini_known_hosts")

	client := &GeminiClient{
		knownHosts: make(map[string]string),
		hostsFile:  hostsFile,
		Timeout:    10 * time.Second,
	}

	client.loadKnownHosts()
	return client
}

func (c *GeminiClient) Fetch(geminiURL string) (*GeminiResponse, error) {
	u, err := url.Parse(geminiURL)
	if err != nil {
		return &GeminiResponse{}, fmt.Errorf("Invalid url: %w\n", err)
	}

	hostname := u.Hostname()
	host := u.Host

	if !strings.Contains(host, ":") {
		host += ":1965"
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		ServerName:         hostname,
	}

	dialer := &net.Dialer{
		Timeout: c.Timeout,
	}

	start := time.Now()
	conn, err := tls.DialWithDialer(dialer, "tcp", host, config)
	if err != nil {
		return &GeminiResponse{}, fmt.Errorf("Connection failed: %v\n", err)
	}

	conn.SetDeadline(start.Add(c.Timeout))
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return &GeminiResponse{}, fmt.Errorf("No certificate received")
	}

	cert := state.PeerCertificates[0]
	fingerprint := getCertFingerprint(cert.Raw)

	if knownFingerprint, exists := c.knownHosts[hostname]; exists {
		if knownFingerprint != fingerprint {
			return &GeminiResponse{}, fmt.Errorf("WARNING: Certificate changed for %s!\nExpected: %s\nGot: %s\n", hostname, knownFingerprint, fingerprint)
		}
	} else {
		// TODO -- answer do you trust or not
		if err := c.saveKnownHost(hostname, fingerprint); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save host: %v\n", err)
		}
	}

	conn.SetWriteDeadline(time.Now().Add(c.Timeout))
	request := geminiURL + "\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		return &GeminiResponse{}, fmt.Errorf("Write failed: %w\n", err)
	}

	conn.SetWriteDeadline(time.Now().Add(c.Timeout))
	reader := bufio.NewReader(conn)
	response := &GeminiResponse{}

	getResponse(response, reader)

	return response, nil
}

func (c *GeminiClient) loadKnownHosts() {
	file, err := os.Open(c.hostsFile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 {
			c.knownHosts[parts[0]] = parts[1]
		}
	}
}

func (c *GeminiClient) saveKnownHost(hostname, fingerprint string) error {
	c.knownHosts[hostname] = fingerprint

	file, err := os.OpenFile(c.hostsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%s %s\n", hostname, fingerprint)
	return err
}

func getCertFingerprint(cert []byte) string {
	hash := sha256.Sum256(cert)
	return hex.EncodeToString(hash[:])
}
