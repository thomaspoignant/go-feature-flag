// Package exporter defines the data exporter of go-feature-flag
//
// These exporters are usable in your init configuration.
//
//
//  ffclient.Init(ffclient.Config{
//    //...
//    DataExporter: ffclient.DataExporter{
//        FlushInterval:   10 * time.Second,
//        MaxEventInMemory: 1000,
//        Exporter: &fileexporter.Exporter{
//            OutputDir: "/output-data/",
//        },
//    },
//    //...
//  })
//
// Check in this package available exporters.
package exporter
