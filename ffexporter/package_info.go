// Package ffexporter defines the data exporter of go-feature-flag
//
// Theses exporters are usable in your init configuration.
//
//
//  ffclient.Init(ffclient.Config{
//    //...
//    DataExporter: ffclient.DataExporter{
//        FlushInterval:   10 * time.Second,
//        MaxEventInMemory: 1000,
//        Exporter: &ffexporter.File{
//            OutputDir: "/output-data/",
//        },
//    },
//    //...
//  })
//
// Check in this package available exporters.
package ffexporter
