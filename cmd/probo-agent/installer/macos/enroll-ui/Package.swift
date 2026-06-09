// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "probo-agent-enroll-ui",
    platforms: [
        .macOS(.v11),
    ],
    targets: [
        .executableTarget(
            name: "probo-agent-enroll-ui",
            path: "Sources"
        ),
    ]
)
