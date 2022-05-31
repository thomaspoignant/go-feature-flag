package ffclient

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)



// KubernetesRetriever is a configuration struct for a Kubernetes retriever.
type KubernetesRetriever struct {
	Namespace     string
	ConfigMapName string
	Key       string
	ClientSet kubernetes.Interface
}

func (s *KubernetesRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	configMap, err := s.ClientSet.CoreV1().ConfigMaps(s.Namespace).Get(ctx, s.ConfigMapName, v1.GetOptions{})
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
