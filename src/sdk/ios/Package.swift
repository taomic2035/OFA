// swift-tools-version:5.7
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "OFA",
    platforms: [
        .iOS(.v15),
        .macOS(.v12),
        .watchOS(.v8)
    ],
    products: [
        .library(
            name: "OFA",
            targets: ["OFA"]
        ),
    ],
    dependencies: [
        // Dependencies can be added here if needed
    ],
    targets: [
        .target(
            name: "OFA",
            dependencies: [],
            path: "OFA",
            exclude: [
                "OFAiOSAgentExample.swift"
            ],
            resources: [],
            publicHeadersPath: nil,
            cSettings: [],
            cppSettings: [],
            swiftSettings: [
                .enableExperimentalFeature("Concurrency")
            ]
        ),
        .testTarget(
            name: "OFATests",
            dependencies: ["OFA"],
            path: "Tests"
        ),
    ],
    swiftLanguageVersions: [.v5]
)
