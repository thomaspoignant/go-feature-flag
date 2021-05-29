package ffclient_test

import (
	"fmt"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

var client *ffclient.GoFeatureFlag

func init() {
	client, _ = ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &ffclient.FileRetriever{Path: "testdata/benchmark/flag-config.yaml"},
	})
}

func BenchmarkBoolVariation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := ffuser.NewUser(fmt.Sprintf("random-key-%d", i))
		_, _ = client.BoolVariation("schedule-flag2", user, false)
	}
}

/* Bemchmark list:

Generate a dynamic flag file in the init method
for all tests.

- boolvariation classic
- boolvariation schedule
- boolvariation progressive
- boolvariation experimentation
- boolvariation 0%
- boolvariation 50%

- idem for all type of variations
- all flag with a lot of flags and different types


*/
