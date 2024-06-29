// swift-tools-version: 5.5
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "go-feature-flag-provider",
    platforms: [
        .iOS(.v14),
        .macOS(.v12)
    ],
    products: [
        .library(
            name: "go-feature-flag-provider",
            targets: ["go-feature-flag-provider"])
    ],
    dependencies: [
        .package(url: "https://github.com/open-feature/swift-sdk.git", from: "0.1.0")
    ],
    targets: [
        .target(
            name: "go-feature-flag-provider",
            dependencies: [
                .product(name: "OpenFeature", package: "swift-sdk")
            ],
            plugins:[]
        ),
        .testTarget(
            name: "go-feature-flag-providerTests",
            dependencies: [
                "go-feature-flag-provider"
            ]
        )
    ]
)
