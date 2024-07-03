// Package exporter defines the data exporter of go-feature-flag
//
// These exporters are usable in your init configuration.
//
//	ffclient.Init(ffclient.Config{
//	  //...
//	  DataExporter: ffclient.DataExporter{
//	   FlushInterval:   10 * time.Second,
//	   MaxEventInMemory: 1000,
//	   DeprecatedExporter: &fileexporter.DeprecatedExporter{
//	     OutputDir: "/output-data/",
//	   },
//	 },
//	 //...
//	})
package exporter
