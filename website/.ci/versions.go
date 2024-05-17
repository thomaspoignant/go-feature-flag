package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	versionFileLocation := "../versions.json"
	versionBakFileLocation := "../versions.json.bak"
	numberVersionToKeep := 8

	content, err := os.ReadFile(versionFileLocation)
	if err != nil {
		log.Fatal("read json:", err)
	}

	var versions []string
	err = json.Unmarshal(content, &versions)
	if err != nil {
		log.Fatal("unmarshal json:", err)
	}

	output := make([]string, 0, numberVersionToKeep)
	for i := 0; i < numberVersionToKeep && i < len(versions); i++ {
		output = append(output, versions[i])
	}

	outputStr, err := json.Marshal(output)
	if err != nil {
		log.Fatal("marshal json:", err)
	}

	err = os.WriteFile(versionFileLocation, outputStr, 0644)
	if err != nil {
		log.Fatal(fmt.Sprintf("write json %s:", versionFileLocation), err)
	}
	err = os.WriteFile(versionBakFileLocation, content, 0644)
	if err != nil {
		log.Fatal(fmt.Sprintf("write json %s:", versionBakFileLocation), err)
	}

}
