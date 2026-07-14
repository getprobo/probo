// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "probo-agent-url-handler",
    platforms: [
        .macOS(.v11),
    ],
    targets: [
        .executableTarget(
            name: "probo-agent-url-handler",
            path: "URLHandlerSources"
        ),
    ]
)
