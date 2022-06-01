package ffclient

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

var kubeClientProvider = func(config *restclient.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

// KubernetesRetriever is a configuration struct for a Kubernetes retriever.
type KubernetesRetriever struct {
	Namespace     string
	ConfigMapName string
	Key           string
	ClientConfig  restclient.Config
	client        kubernetes.Interface
}

func (s *KubernetesRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if s.client == nil {
		client, clientErr := kubeClientProvider(&s.ClientConfig)
		if clientErr != nil {
			return nil, fmt.Errorf("unable to create client, error: %s", clientErr)
		}
		s.client = client
	}
	configMap, err := s.client.CoreV1().ConfigMaps(s.Namespace).Get(ctx, s.ConfigMapName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read from config map %s.%s, error: %s", s.ConfigMapName, s.Namespace, err,
		)
	}
	content, ok := configMap.Data[s.Key]
	if !ok {
		return nil, fmt.Errorf("key %s not existing in config map %s.%s", s.Key, s.ConfigMapName, s.Namespace)
	}
	return []byte(content), nil
}
