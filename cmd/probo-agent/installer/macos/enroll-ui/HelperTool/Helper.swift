import Foundation
import ProboAgentShared
import os

final class Helper: NSObject, ProboAgentHelperProtocol, NSXPCListenerDelegate {
    private static let log = Logger(
        subsystem: "com.probo.agent.helper",
        category: "Helper"
    )

    func getVersion(withReply reply: @escaping (String) -> Void) {
        reply(ProboAgentHelperConstants.helperVersion)
    }

    func ping(withReply reply: @escaping (Bool) -> Void) {
        Self.log.info("ping")
        reply(true)
    }

    func install(
        serverURL: String,
        enrollmentToken: String,
        configDir: String,
        withReply reply: @escaping (Int32, String?) -> Void
    ) {
        let trimmedServer = serverURL.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedToken = enrollmentToken.trimmingCharacters(in: .whitespacesAndNewlines)
        let dir = configDir.trimmingCharacters(in: .whitespacesAndNewlines)

        guard !trimmedServer.isEmpty, !trimmedToken.isEmpty else {
            reply(1, "server URL and enrollment token are required")
            return
        }

        // Pass the token via env rather than argv so it does not show up in
        // process listings (ps / Activity Monitor). Install already accepts
        // PROBO_ENROLLMENT_TOKEN when --enrollment-token is omitted.
        var args = [
            "install",
            "--server", trimmedServer,
        ]
        if !dir.isEmpty {
            args.append(contentsOf: ["--dir", dir])
        }

        var environment = ProcessInfo.processInfo.environment
        environment["PROBO_ENROLLMENT_TOKEN"] = trimmedToken

        let result = runAgent(args: args, environment: environment)
        reply(result.exitCode, result.output)
    }

    func listener(_ listener: NSXPCListener, shouldAcceptNewConnection connection: NSXPCConnection)
        -> Bool
    {
        guard ClientAuth.accepts(connection: connection) else {
            Self.log.error(
                "refused XPC connection pid=\(connection.processIdentifier, privacy: .public)")
            return false
        }

        Self.log.info(
            "accepted XPC connection pid=\(connection.processIdentifier, privacy: .public)")
        connection.exportedInterface = NSXPCInterface(with: ProboAgentHelperProtocol.self)
        connection.exportedObject = self
        connection.resume()
        return true
    }

    private struct CommandResult {
        let exitCode: Int32
        let output: String?
    }

    private func runAgent(args: [String], environment: [String: String]? = nil) -> CommandResult {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: ProboAgentHelperConstants.agentExecutablePath)
        process.arguments = args
        if let environment {
            process.environment = environment
        }

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = pipe

        do {
            try process.run()
            let data = try pipe.fileHandleForReading.readToEnd() ?? Data()
            process.waitUntilExit()
            let text = String(data: data, encoding: .utf8)?
                .trimmingCharacters(in: .whitespacesAndNewlines)

            return CommandResult(
                exitCode: process.terminationStatus,
                output: text?.isEmpty == false ? text : nil
            )
        } catch {
            return CommandResult(exitCode: 1, output: error.localizedDescription)
        }
    }

}
