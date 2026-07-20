// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "probo-agent-url-handler",
    platforms: [
        .macOS(.v11)
    ],
    products: [
        .library(name: "ProboAgentShared", targets: ["ProboAgentShared"]),
        .library(name: "HelperClient", targets: ["HelperClient"]),
        .executable(name: "com.probo.agent.helper", targets: ["com.probo.agent.helper"]),
        .executable(name: "probo-agent-url-handler", targets: ["probo-agent-url-handler"]),
    ],
    targets: [
        .target(
            name: "ProboAgentShared",
            path: "Shared",
            exclude: [
                "HelperVersion.generated.swift.tmpl",
                "SigningConstants.generated.swift.tmpl",
            ],
            linkerSettings: [
                .linkedFramework("Security")
            ]
        ),
        .target(
            name: "HelperClient",
            dependencies: ["ProboAgentShared"],
            path: "HelperClient"
        ),
        .executableTarget(
            name: "com.probo.agent.helper",
            dependencies: ["ProboAgentShared"],
            path: "HelperTool",
            exclude: ["Info.plist.tmpl", "Launchd.plist.tmpl"],
            linkerSettings: [
                .linkedFramework("Security")
            ]
        ),
        .executableTarget(
            name: "probo-agent-url-handler",
            dependencies: ["HelperClient", "ProboAgentShared"],
            path: "URLHandlerSources"
        ),
    ]
)
