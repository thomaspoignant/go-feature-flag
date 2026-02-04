package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type Versions struct {
	Maven Maven `json:"maven,omitempty"`
	Npm   Npm   `json:"npm,omitempty"`
	Pypi  Pypi  `json:"pypi,omitempty"`
	Nuget Nuget `json:"nuget,omitempty"`
	GO    GO    `json:"go,omitempty"`
	Swift Swift `json:"swift,omitempty"`
}
type Maven struct {
	Sdk            string `json:"sdk,omitempty"`
	KotlinProvider string `json:"providerKt,omitempty"`
	JavaProvider   string `json:"providerJava,omitempty"`
	Android        string `json:"android,omitempty"`
	KotlinSDK      string `json:"kotlinSdk,omitempty"`
}

type Npm struct {
	Core           string `json:"core,omitempty"`
	ServerSDK      string `json:"serverSDK,omitempty"`
	WebSDK         string `json:"webSDK,omitempty"`
	ServerProvider string `json:"providerServer,omitempty"`
	WebProvider    string `json:"providerWeb,omitempty"`
}

type Pypi struct {
	OpenFeatureSDK string `json:"sdk,omitempty"`
	PythonProvider string `json:"provider,omitempty"`
}

type Nuget struct {
	OpenFeatureSDK string `json:"sdk,omitempty"`
	Provider       string `json:"provider,omitempty"`
}

type Swift struct {
	Provider string `json:"provider,omitempty"`
}

type GO struct {
	Provider string `json:"provider,omitempty"`
	SDK      string `json:"sdk,omitempty"`
}

func main() {
	var wg sync.WaitGroup
	versions := Versions{}
	wg.Add(16)
	go func() {
		defer wg.Done()
		versions.Swift.Provider = getSwiftVersion("go-feature-flag/openfeature-swift-provider")
	}()
	go func() {
		defer wg.Done()
		versions.Maven.Android = getMavenVersion("dev.openfeature", "kotlin-sdk")
	}()
	go func() {
		defer wg.Done()
		versions.Maven.KotlinSDK = getMavenVersion("dev.openfeature", "android-sdk")
	}()
	go func() {
		defer wg.Done()
		versions.Maven.Sdk = getMavenVersion("dev.openfeature", "sdk")
	}()
	go func() {
		defer wg.Done()
		versions.Maven.KotlinProvider = getMavenVersion("org.gofeatureflag.openfeature", "gofeatureflag-kotlin-provider")
	}()
	go func() {
		defer wg.Done()
		versions.Maven.JavaProvider = getMavenVersion("dev.openfeature.contrib.providers", "go-feature-flag")
	}()

	go func() {
		defer wg.Done()
		versions.Npm.Core = getNPMVersion("@openfeature/core")
	}()

	go func() {
		defer wg.Done()
		versions.Npm.ServerSDK = getNPMVersion("@openfeature/server-sdk")
	}()

	go func() {
		defer wg.Done()
		versions.Npm.WebSDK = getNPMVersion("@openfeature/web-sdk")
	}()

	go func() {
		defer wg.Done()
		versions.Npm.ServerProvider = getNPMVersion("@openfeature/go-feature-flag-provider")
	}()

	go func() {
		defer wg.Done()
		versions.Npm.WebProvider = getNPMVersion("@openfeature/go-feature-flag-web-provider")
	}()

	go func() {
		defer wg.Done()
		versions.Pypi.OpenFeatureSDK = getPypiVersion("openfeature-sdk")
	}()

	go func() {
		defer wg.Done()
		versions.Pypi.PythonProvider = getPypiVersion("gofeatureflag-python-provider")
	}()
	go func() {
		defer wg.Done()
		versions.Nuget.OpenFeatureSDK = getNugetVersion("OpenFeature")
	}()
	go func() {
		defer wg.Done()
		versions.Nuget.Provider = getNugetVersion("OpenFeature.Contrib.GOFeatureFlag")
	}()
	go func() {
		defer wg.Done()
		versions.GO.Provider = getGOVersion("github.com/open-feature/go-sdk-contrib/providers/go-feature-flag")
	}()
	go func() {
		defer wg.Done()
		versions.GO.SDK = getGOVersion("github.com/open-feature/go-sdk")
	}()
	wg.Wait()

	var content []byte
	content, err := json.Marshal(versions)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("./website/static/sdk-versions.json", content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getMavenVersion(groupId string, artifactId string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "search.maven.org",
		Path:   "solrsearch/select",
	}
	q := u.Query()
	q.Set("q", fmt.Sprintf("g:\"%s\" AND a:\"%s\"", groupId, artifactId))
	q.Set("core", "gav")
	q.Set("rows", "1")
	q.Set("wt", "json")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type mavenRes struct {
		Response struct {
			Docs []struct {
				V string `json:"v,omitempty"`
			} `json:"docs,omitempty"`
		} `json:"response,omitempty"`
	}

	var res mavenRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}

	return res.Response.Docs[0].V
}

func getNPMVersion(packageName string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "registry.npmjs.org",
		Path:   packageName,
	}
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type npmRes struct {
		DistTags struct {
			Latest string `json:"latest,omitempty"`
		} `json:"dist-tags,omitempty"`
	}

	var res npmRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}

	return res.DistTags.Latest
}

func getPypiVersion(packageName string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "pypi.org",
		Path:   fmt.Sprintf("/pypi/%s/json", packageName),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type pypiRes struct {
		Info struct {
			Version string `json:"version,omitempty"`
		} `json:"info,omitempty"`
	}

	var res pypiRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}

	return res.Info.Version
}

func getNugetVersion(packageName string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "api.nuget.org",
		Path:   fmt.Sprintf("v3-flatcontainer/%s/index.json", packageName),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type nugetRes struct {
		Versions []string `json:"versions,omitempty"`
	}

	var res nugetRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}
	return res.Versions[len(res.Versions)-1]
}

func getGOVersion(packageName string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "proxy.golang.org",
		Path:   fmt.Sprintf("%s/@latest", packageName),
	}

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type goRes struct {
		Version string `json:"version,omitempty"`
	}

	var res goRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}
	return res.Version
}

func getSwiftVersion(repoSlug string) string {
	//
	u := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("repos/%s/releases/latest", repoSlug),
	}

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type githubRes struct {
		TagName string `json:"tag_name,omitempty"`
	}

	var res githubRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Fatal(err)
	}

	return res.TagName
}
