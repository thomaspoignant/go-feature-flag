package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"os"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffexporter"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func main() {
	/*
		1. You need to have your AWS credentials in your machine to use this example.
		2. Create a bucket in your account, and replace the bucket name "my-test-bucket" in the configuration
		3. Run this example
		4. In your bucket you will have 2 files:
			   - /go-feature-flag/variations/flag-variation-EXAMPLE-<timestamp>.json
			   - /go-feature-flag/variations/flag-variation-EXAMPLE-<timestamp>.json
	*/
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Logger:          log.New(os.Stdout, "", 0),
		Context:         context.Background(),
		Retriever: &ffclient.FileRetriever{
			Path: "examples/data_export_s3/flags.yaml",
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    1 * time.Second,
			MaxEventInMemory: 100,
			Exporter: &ffexporter.S3{
				Format:   "json",
				Bucket:   "my-test-bucket",
				S3Path:   "/go-feature-flag/variations/",
				Filename: "flag-variation-{{ .Timestamp}}.{{ .Format}}",
				AwsConfig: &aws.Config{
					Region: aws.String("eu-west-1"),
				},
			},
		},
	})

	// Check init errors.
	if err != nil {
		log.Fatal(err)
	}
	// defer closing ffclient
	defer ffclient.Close()

	// create users
	user1 := ffuser.NewAnonymousUser("aea2fdc1-b9a0-417a-b707-0c9083de68e3")
	user2 := ffuser.NewUser("332460b9-a8aa-4f7a-bc5d-9cc33632df9a")

	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)
	_, _ = ffclient.StringVariation("unknown-flag", user1, "defaultValue")
	_, _ = ffclient.JSONVariation("unknown-flag-2", user1, map[string]interface{}{"test": "toto"})

	// Wait 2 seconds to have a second file
	time.Sleep(2 * time.Second)
	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)

	/*
		The content of those files should looks like:
		/go-feature-flag/variations/flag-variation-EXAMPLE-<timestamp>.json:
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"new-admin-access","variation":"True","value":true,"default":false}
			{"kind":"feature","contextKind":"user","userKey":"332460b9-a8aa-4f7a-bc5d-9cc33632df9a","creationDate":1618234129,"key":"new-admin-access","variation":"False","value":false,"default":false}
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"unknown-flag","variation":"SdkDefault","value":"defaultValue","default":true}
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"unknown-flag-2","variation":"SdkDefault","value":{"test":"toto"},"default":true}
		----
		/go-feature-flag/variations/flag-variation-EXAMPLE-<timestamp>.json:
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234131,"key":"new-admin-access","variation":"True","value":true,"default":false}
			{"kind":"feature","contextKind":"user","userKey":"332460b9-a8aa-4f7a-bc5d-9cc33632df9a","creationDate":1618234131,"key":"new-admin-access","variation":"False","value":false,"default":false}
	*/
}
