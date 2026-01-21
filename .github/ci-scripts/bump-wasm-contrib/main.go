package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// This script allows to bump the wasm version in the different contrib repositories.
// It is used to update the wasm version in the different contrib repositories when
// a new version of the wasm is released.
// It is used to bump all the repositories at once.
//
// All contrib repositories must have been checked out in the current directory.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bump-wasm-contrib <wasm-version>")
		os.Exit(1)
	}
	wasmVersion := os.Args[1]
	outDir := "./out/contrib"

	// Java: Bump the wasm version in the pom.xml file
	replaceLine(
		fmt.Sprintf("%s/java-sdk-contrib/providers/go-feature-flag/pom.xml", outDir),
		fmt.Sprintf(`        <wasm.version>%s</wasm.version>`, wasmVersion),
		"<wasm.version>",
	)

	// .NET: Bump the wasm version in the OpenFeature.Providers.GOFeatureFlag.csproj file
	replaceLine(
		fmt.Sprintf("%s/dotnet-sdk-contrib/src/OpenFeature.Providers.GOFeatureFlag/OpenFeature.Providers.GOFeatureFlag.csproj", outDir),
		fmt.Sprintf("    <WasmVersion>%s</WasmVersion>", wasmVersion),
		"<WasmVersion>",
	)

	// JavaScript: Bump the wasm version in the config.js file
	replaceLine(
		fmt.Sprintf("%s/js-sdk-contrib/libs/providers/go-feature-flag/scripts/copy-latest-wasm.js", outDir),
		fmt.Sprintf(`const TARGET_WASM_VERSION = "%s";`, wasmVersion),
		"const TARGET_WASM_VERSION",
	)
}

// bumpWasmJavascript allows to bump the wasm version in a JavaScript file
// inputFile: the JavaScript file to update
// newValue: the new value to set (e.g. "v2.0.0")
func replaceLine(inputFile, newValue, lineMatcher string) {
	file, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line defines our variable
		// Using TrimSpace handles indentation
		if strings.HasPrefix(strings.TrimSpace(line), lineMatcher) {
			// Replace the whole line
			lines = append(lines, newValue)
		} else {
			// Keep original line
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	os.WriteFile(inputFile, []byte(strings.Join(lines, "\n")), 0644)
	fmt.Printf("%s updated successfully.\n", inputFile)
}
