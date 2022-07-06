// Package ffclient aids adding instrumentation to have feature flags in your
// app without any backend server.
//
// Summary
//
// This package and its subpackages contain bits of code to have an easy feature
// flag solution with no complex installation to do on your infrastructure and
// without using 3rd party vendor for this.
//
// The ffclient package provides the entry point - initialization and the basic
// method to get your flags value.
//
// Before using the module you need to initialized it this way:
//  func main() {
//    err := ffclient.Init(ffclient.Config{
//             PollingInterval: 3 * time.Second,
//             HTTPRetriever: &httpretriever.Retriever{
//               URL:    "http://example.com/flag-config.yaml",
//             },
//           })
//    defer ffclient.Close()
//    ...
// This example will load a file from an HTTP endpoint and will refresh the flags every 3 seconds.
//
// Now you can evalute your flags anywhere in your code.
//
//  func main() {
//    ...
//    user := ffuser.NewUser("user-unique-key")
//    hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
//    if hasFlag {
//      //flag "test-flag" is true for the user
//    } else {
//      // flag "test-flag" is false for the user
//    }
//    ...
package ffclient
