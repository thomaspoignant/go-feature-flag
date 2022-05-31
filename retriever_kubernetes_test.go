package ffclient

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

var expectedContent = `test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false

test-flag2:
  rule: key eq "not-a-key"
  percentage: 100
  true: true
  false: false
  default: false
`

func Test_kubernetesRetriever_Retrieve(t *testing.T) {
	type fields struct {
		object        runtime.Object
		namespace     string
		configMapName string
		key           string
		context       context.Context
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
					Data: map[string]string{"valid": expectedContent},
				},
				namespace:     "Namespace",
				configMapName: "ConfigMap1",
				key:           "valid",
			},
			want:[]byte(expectedContent),
			wantErr: nil,
		},
		{
			name: "Key not existing",
			fields: fields{
				object: &api.ConfigMap{
					ObjectMeta: v1.ObjectMeta{Name: "ConfigMap1", Namespace: "Namespace"},
					Data: map[string]string{"valid": expectedContent},
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
					Data: map[string]string{"valid": expectedContent},
				},
				namespace:     "WrongNamespace",
				configMapName: "NotExisting",
			},
			wantErr: errors.New("unable to read from config map NotExisting.WrongNamespace, error: configmaps \"NotExisting\" not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := KubernetesRetriever{
				ClientSet:     fake.NewSimpleClientset(tt.fields.object),
				ConfigMapName: tt.fields.configMapName,
				Key:           tt.fields.key,
				Namespace: tt.fields.namespace,
			}
			got, err := s.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				assert.Equal(t, tt.want, got, "Retrieve() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
