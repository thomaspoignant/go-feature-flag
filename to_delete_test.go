package ffclient_test

import (
	"testing"

	"github.com/thomaspoignant/go-feature-flag/modules/core"
	"github.com/thomaspoignant/go-feature-flag/modules/evaluation"
)

func TestToDelete(t *testing.T) {
	core.Core()
	evaluation.Evaluation()
}
