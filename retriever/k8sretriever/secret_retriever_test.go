package k8sretriever

import (
	"context"
	"encoding/base64"
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

var expectedDecodedContent = `
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

var expectedEncodedContent = base64.StdEncoding.EncodeToString([]byte(expectedDecodedContent))

func Test_kubernetesSecretRetriever_Retrieve(t *testing.T) {
	originalKubeClientProvider := kubeClientProvider
	defer func() {
		kubeClientProvider = originalKubeClientProvider
	}()

	kubeClientProviderFactory := func(object ...runtime.Object) func(*restclient.Config) (kubernetes.Interface, error) {
		return func(config *restclient.Config) (kubernetes.Interface, error) {
			return fake.NewClientset(object...), nil
		}
	}

	type fields struct {
		object     runtime.Object
		namespace  string
		secretName string
		secretKey  string
		context    context.Context
		setClient  bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr error
	}{
		{
			name: "Secret existing",
			fields: fields{
				object: &api.Secret{
					ObjectMeta: v1.ObjectMeta{Name: "Secret1", Namespace: "Namespace"},
					StringData: map[string]string{"FEATURE_FLAGS": expectedEncodedContent},
				},
				namespace:  "Namespace",
				secretName: "Secret1",
				secretKey:  "FEATURE_FLAGS",
			},
			want:    expectedDecodedContent,
			wantErr: nil,
		},
		{
			name: "Key not existing",
			fields: fields{
				object: &api.Secret{
					ObjectMeta: v1.ObjectMeta{Name: "Secret1", Namespace: "Namespace"},
					StringData: map[string]string{"FEATURE_FLAGS": expectedEncodedContent},
				},
				namespace:  "Namespace",
				secretName: "Secret1",
				secretKey:  "INVALID",
			},
			wantErr: errors.New("key INVALID not existing in secret Secret1.Namespace"),
		},
		{
			name: "Config Map not existing",
			fields: fields{
				object: &api.Secret{
					ObjectMeta: v1.ObjectMeta{Name: "Secret1", Namespace: "Namespace"},
					Data:       map[string][]byte{"FEATURE_FLAGS": []byte(expectedEncodedContent)},
				},
				namespace:  "WrongNamespace",
				secretName: "NotExisting",
			},
			wantErr: errors.New(
				"unable to read from secret NotExisting.WrongNamespace, error: secrets \"NotExisting\" not found",
			),
		},
		{
			name: "Client already there",
			fields: fields{
				object: &api.Secret{
					ObjectMeta: v1.ObjectMeta{Name: "Secret1", Namespace: "Namespace"},
					StringData: map[string]string{"FEATURE_FLAGS": expectedEncodedContent},
				},
				namespace:  "Namespace",
				secretName: "Secret1",
				secretKey:  "FEATURE_FLAGS",
				setClient:  true,
			},
			wantErr: errors.New(
				"unable to read from secret Secret1.Namespace, error: secrets \"Secret1\" not found",
			),
		},
		{
			name: "Secret decoding fails",
			fields: fields{
				object: &api.Secret{
					ObjectMeta: v1.ObjectMeta{Name: "Secret1", Namespace: "Namespace"},
					StringData: map[string]string{"FEATURE_FLAGS": "_INVALID"},
				},
				namespace:  "Namespace",
				secretName: "Secret1",
				secretKey:  "FEATURE_FLAGS",
			},
			wantErr: errors.New(
				"unable to decode secret Secret1.Namespace, error: illegal base64 data at input byte 0",
			),
		},
		{
			name: "k8s client is nil",
			fields: fields{
				object: &api.Secret{
					ObjectMeta: v1.ObjectMeta{Name: "Secret1", Namespace: "Namespace"},
					StringData: map[string]string{"FEATURE_FLAGS": expectedEncodedContent},
				},
				namespace:  "Namespace",
				secretName: "Secret1",
				secretKey:  "FEATURE_FLAGS",
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
			s := SecretRetriever{
				SecretName: tt.fields.secretName,
				SecretKey:  tt.fields.secretKey,
				Namespace:  tt.fields.namespace,
			}
			if tt.fields.setClient {
				s.client = fake.NewClientset()
			}
			got, err := s.Retrieve(tt.fields.context)

			assert.Equal(t, tt.wantErr, err, "retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				assert.Equal(
					t,
					tt.want,
					string(got),
					"retrieve() got = %v, want %v",
					string(got),
					tt.want,
				)
			}
		})
	}
}
