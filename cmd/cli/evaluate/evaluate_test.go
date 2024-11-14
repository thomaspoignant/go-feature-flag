package evaluate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_evaluate_Evaluate(t *testing.T) {
	type fields struct {
		config        string
		fileFormat    string
		flag          string
		evaluationCtx string
	}
	tests := []struct {
		name     string
		evaluate evaluate
		wantErr  assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval := tt.evaluate.Evaluate()
			tt.wantErr(t, eval)

			if err := tt.evaluate.Evaluate(); (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
