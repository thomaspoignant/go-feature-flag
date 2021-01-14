package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func main() {
	// Init ffclient with a file retriever.
	err := ffclient.Init(ffclient.Config{
		PollInterval: 10,
		Logger:       log.New(os.Stdout, "", 0),
		Context:      context.Background(),
		Retriever: &ffclient.FileRetriever{
			Path:     "examples/file/flags.yaml",
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
	user3 := ffuser.NewUserBuilder("785a14bf-d2c5-4caa-9c70-2bbc4e3732a5").
		AddCustom("email", "user2@email.com").
		AddCustom("firstname", "John").
		AddCustom("lastname", "Doe").
		AddCustom("admin", true).
		Build()

	// --- test flag with no rule
	// user1
	user1HasAccessToNewAdmin, err := ffclient.BoolVariation("new-admin-access", user1, false)
	if err != nil {
		// we log the error, but we still have a meaningful value in user1HasAccessToNewAdmin (the default value).
		log.Printf("something went wrong when getting the flag: %v", err)
	}
	if user1HasAccessToNewAdmin {
		fmt.Println("user1 has access to the new admin")
	}

	// user2
	user2HasAccessToNewAdmin, err := ffclient.BoolVariation("new-admin-access", user2, false)
	if err != nil {
		// we log the error, but we still have a meaningful value in hasAccessToNewAdmin (the default value).
		log.Printf("something went wrong when getting the flag: %v", err)
	}
	if !user2HasAccessToNewAdmin {
		fmt.Println("user2 has not access to the new admin")
	}

	// --- test flag with rule only for admins
	// user 1 is not admin so should not access to the flag
	user1HasAccess, _ := ffclient.BoolVariation("flag-only-for-admin", user1, false)
	if !user1HasAccess {
		fmt.Println("user1 is not admin so no access to the flag")
	}

	// user 3 is admin and the flag apply to this key.
	user3HasAccess, _ := ffclient.BoolVariation("flag-only-for-admin", user3, false)
	if user3HasAccess {
		fmt.Println("user 3 is admin and the flag apply to this key.")
	}
}
