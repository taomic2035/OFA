// Package.swift
// swift-tools-version:5.9
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "OFAAgent",
    platforms: [
        .iOS(.v13),
        .macOS(.v12)
    ],
    products: [
        .library(
            name: "OFAAgent",
            targets: ["OFAAgent"]),
    ],
    dependencies: [
        .package(url: "https://github.com/grpc/grpc-swift.git", from: "1.20.0"),
        .package(url: "https://github.com/apple/swift-protobuf.git", from: "1.25.0"),
    ],
    targets: [
        .target(
            name: "OFAAgent",
            dependencies: [
                .product(name: "GRPC", package: "grpc-swift"),
                .product(name: "NIOCore", package: "grpc-swift"),
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
            ],
            path: "Sources/OFAAgent"),
        .testTarget(
            name: "OFAAgentTests",
            dependencies: ["OFAAgent"],
            path: "Tests/OFAAgentTests"),
    ]
)