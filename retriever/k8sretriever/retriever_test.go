package k8sretriever

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	api "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
)

var expectedContent = `
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var	
  trackEvents: false

test-flag2:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "not-a-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var	
  trackEvents: false
`

func Test_kubernetesRetriever_Retrieve(t *testing.T) {
	originalKubeClientProvider := kubeClientProvider
	defer func() {
		kubeClientProvider = originalKubeClientProvider
	}()

	kubeClientProviderFactory := func(object ...runtime.Object) func(*restclient.Config) (kubernetes.Interface, error) {
		return func(config *restclient.Config) (kubernetes.Interface, error) {
			return fake.NewSimpleClientset(object...), nil
		}
	}

	type fields struct {
		object        runtime.Object
		namespace     string
		configMapName string
		key           string
		context       context.Context
		setClient     bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr error
	}{
		{
			name: "ConfigMap existing",
			fields: fields{
				object: &api.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "ConfigMap1", Namespace: "Namespace"},
					Data:       map[string]string{"valid": expectedContent},
				},
				namespace:     "Namespace",
				configMapName: "ConfigMap1",
				key:           "valid",
			},
			want:    []byte(expectedContent),
			wantErr: nil,
		},
		{
			name: "Key not existing",
			fields: fields{
				object: &api.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "ConfigMap1", Namespace: "Namespace"},
					Data:       map[string]string{"valid": expectedContent},
				},
				namespace:     "Namespace",
				configMapName: "ConfigMap1",
				key:           "INVALID",
			},
			wantErr: errors.New("key INVALID not existing in config map ConfigMap1.Namespace"),
		},
		{
			name: "Config Map not existing",
			fields: fields{
				object: &api.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "ConfigMap1", Namespace: "Namespace"},
					Data:       map[string]string{"valid": expectedContent},
				},
				namespace:     "WrongNamespace",
				configMapName: "NotExisting",
			},
			wantErr: errors.New("unable to read from config map NotExisting.WrongNamespace, error: configmaps \"NotExisting\" not found"),
		},
		{
			name: "Client already there",
			fields: fields{
				object: &api.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "ConfigMap1", Namespace: "Namespace"},
					Data:       map[string]string{"valid": expectedContent},
				},
				namespace:     "Namespace",
				configMapName: "ConfigMap1",
				key:           "valid",
				setClient:     true,
			},
			wantErr: errors.New("unable to read from config map ConfigMap1.Namespace, error: configmaps \"ConfigMap1\" not found"),
		},
		{
			name: "k8s client is nil",
			fields: fields{
				object: &api.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "ConfigMap1", Namespace: "Namespace"},
					Data:       map[string]string{"valid": expectedContent},
				},
				namespace:     "Namespace",
				configMapName: "ConfigMap1",
				key:           "valid",
			},
			wantErr: errors.New("k8s client is nil after initialization"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClientProvider = kubeClientProviderFactory(tt.fields.object)
			if tt.name == "k8s client is nil" {
				// mocking the kubeClientProvider function
				kubeClientProvider = func(config *restclient.Config) (kubernetes.Interface, error) {
					return nil, nil
				}
			}
			s := Retriever{
				ConfigMapName: tt.fields.configMapName,
				Key:           tt.fields.key,
				Namespace:     tt.fields.namespace,
			}
			if tt.fields.setClient {
				s.client = fake.NewSimpleClientset()
			}
			got, err := s.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				assert.Equal(t, tt.want, got, "Retrieve() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
