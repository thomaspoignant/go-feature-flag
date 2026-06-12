package httpretriever

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/retriever/shared"
)

// Retriever is a configuration struct for an HTTP endpoint retriever.
type Retriever struct {
	// URL of your endpoint
	URL string

	// HTTP Method we should use (default: GET)
	Method string

	// Body of the request if needed (default: empty body)
	Body string

	// Header added to the request
	Header http.Header

	// Timeout we should wait before failing (default: 10 seconds)
	Timeout time.Duration

	// ClientCertPath is the path to the client certificate file used for mTLS.
	ClientCertPath string

	// ClientKeyPath is the path to the client certificate key file used for mTLS.
	ClientKeyPath string

	// CACertPath is the path to the CA certificate file used to verify the server certificate.
	CACertPath string

	httpClient internal.HTTPClient
}

// SetHTTPClient is here if you want to override the default http.Client we are using.
// It is also used for the tests.
func (r *Retriever) SetHTTPClient(client internal.HTTPClient) {
	r.httpClient = client
}

// Retrieve is the function in charge of fetching the flag configuration.
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	httpClient, err := r.getHTTPClient()
	if err != nil {
		return nil, err
	}

	resp, err := shared.CallHTTPAPI(ctx, r.URL, r.Method, r.Body, r.Timeout, r.Header, httpClient)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("request to %s failed with code %d", r.URL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Retriever) getHTTPClient() (internal.HTTPClient, error) {
	if r.httpClient != nil {
		return r.httpClient, nil
	}

	var transport http.RoundTripper
	if r.ClientCertPath != "" || r.ClientKeyPath != "" || r.CACertPath != "" {
		tlsConfig, err := r.tlsConfig()
		if err != nil {
			return nil, err
		}

		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("http default transport is %T, expected *http.Transport", http.DefaultTransport)
		}

		tlsTransport := defaultTransport.Clone()
		tlsTransport.TLSClientConfig = tlsConfig
		transport = tlsTransport
	}

	r.httpClient = internal.NewHTTPClient(r.Timeout, transport)
	return r.httpClient, nil
}

func (r *Retriever) tlsConfig() (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	certificates, err := r.clientCertificates()
	if err != nil {
		return nil, err
	}
	config.Certificates = certificates

	rootCAs, err := r.rootCAs()
	if err != nil {
		return nil, err
	}
	config.RootCAs = rootCAs

	return config, nil
}

func (r *Retriever) clientCertificates() ([]tls.Certificate, error) {
	if (r.ClientCertPath == "") != (r.ClientKeyPath == "") {
		return nil, fmt.Errorf("client certificate and client key must be provided together")
	}
	if r.ClientCertPath == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(r.ClientCertPath, r.ClientKeyPath)
	if err != nil {
		return nil, err
	}
	return []tls.Certificate{cert}, nil
}

func (r *Retriever) rootCAs() (*x509.CertPool, error) {
	if r.CACertPath == "" {
		return nil, nil
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	caCert, err := os.ReadFile(r.CACertPath)
	if err != nil {
		return nil, err
	}
	if !rootCAs.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}
	return rootCAs, nil
}
