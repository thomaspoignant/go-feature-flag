package ffclient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/modules/core"
	"github.com/thomaspoignant/go-feature-flag/modules/evaluation"
)

func TestToDelete(t *testing.T) {
	assert.Panics(t, func() { core.Core() }, "core.Core() should panic")
	assert.Panics(t, func() { evaluation.Evaluation() }, "evaluation.Evaluation() should panic")
}
