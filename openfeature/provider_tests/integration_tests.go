package main

import (
	"bytes"
	"fmt"
	"github.com/bitfield/script"
	"log"
	"net/http"
	"time"
)

const portRelayProxy = "1031"
const timeoutRelayProxyReady = 1 * time.Minute
const containerNameRelayProxy = "relayproxyintegrationtests"

var testCommands = []string{"ls"}

func main() {
	relayProxyConfigFile := "/Users/thomas.poignant/dev/thomaspoignant/go-feature-flag/cmd/relayproxy/testdata/config/valid-file.yaml"
	// Stop to ensure that it was not running before
	stopRelayProxy()

	// Start the relay proxy and wait to be ready
	startRelayProxy(relayProxyConfigFile)
	defer stopRelayProxy()
	waitForRelayProxy()

	for _, cmd := range testCommands {
		var output bytes.Buffer
		p := script.Exec(cmd).WithStdout(&output)
		p.Stdout()
		if p.ExitStatus() != 0 {
			log.Fatalf("error while launching command: %s\n %s\n, Error:%s", cmd, output.String(), p.Error())
		} else {
			log.Printf("launching command: %s\n%s\n", cmd, output.String())
		}
	}

	fmt.Println("Relay proxy is ready")
}

func startRelayProxy(relayProxyConfigFile string) {
	var output bytes.Buffer
	startDocker := "docker run --name " + containerNameRelayProxy + " -d -p " + portRelayProxy + ":" + portRelayProxy + " -v " +
		relayProxyConfigFile +
		":/goff/goff-proxy.yaml thomaspoignant/go-feature-flag-relay-proxy:latest"

	script.Exec("docker stop " + containerNameRelayProxy).Wait()
	script.Exec("docker rm " + containerNameRelayProxy).Wait()
	p := script.Exec(startDocker).WithStdout(&output)
	p.Stdout()
	if p.ExitStatus() != 0 || p.Error() != nil {
		panic(fmt.Sprintf("impossible to start the relay proxy, err:%s", output.String()))
	}
}

func stopRelayProxy() {
	p := script.Exec("docker stop " + containerNameRelayProxy).Exec("docker rm " + containerNameRelayProxy)
	p.Wait()
	fmt.Println("Relay proxy is stopped")
}

// waitForRelayProxy is calling the health API until the relay proxy is up and running.
func waitForRelayProxy() {
	maxWaitTime := time.Now().Add(timeoutRelayProxyReady)
	for time.Now().Before(maxWaitTime) {
		resp, _ := http.Get("http://localhost:" + portRelayProxy + "/health")
		if resp != nil && resp.StatusCode == http.StatusOK {
			return
		}
		fmt.Println("waiting relay proxy to be ready")
		time.Sleep(1 * time.Second)
	}
	panic("Timeout while waiting for relay proxy to be ready")
}
