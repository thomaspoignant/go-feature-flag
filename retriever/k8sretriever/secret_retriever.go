package k8sretriever

import (
	"context"
	"encoding/base64"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// SecretRetriever is a configuration struct for a Kubernetes Secret retriever.
type SecretRetriever struct {
	Namespace    string
	SecretName   string
	SecretKey    string
	ClientConfig restclient.Config
	client       kubernetes.Interface
}

// Retrieve is the function in charge of fetching the flag configuration.
func (s *SecretRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if s.client == nil {
		client, clientErr := kubeClientProvider(&s.ClientConfig)
		if clientErr != nil {
			return nil, fmt.Errorf("unable to create client, error: %s", clientErr)
		}
		s.client = client
	}

	if s.client == nil {
		return nil, fmt.Errorf("k8s client is nil after initialization")
	}

	secret, err := s.client.CoreV1().Secrets(s.Namespace).Get(ctx, s.SecretName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read from secret %s.%s, error: %s", s.SecretName, s.Namespace, err,
		)
	}

	encodedContent, ok := secret.StringData[s.SecretKey]
	if !ok {
		return nil, fmt.Errorf(
			"key %s not existing in secret %s.%s",
			s.SecretKey,
			s.SecretName,
			s.Namespace,
		)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(encodedContent)
	if err != nil {
		return nil, fmt.Errorf("unable to decode secret %s.%s, error: %s", s.SecretName, s.Namespace, err)
	}

	return decodedContent, nil
}
