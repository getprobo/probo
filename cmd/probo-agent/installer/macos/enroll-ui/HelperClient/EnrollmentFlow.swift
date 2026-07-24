import Foundation
import ProboAgentShared

public struct EnrollPreflightResult: Decodable {
    public let server: String
    public let token: String
    public let alreadyEnrolled: Bool
    public let configDir: String
}

public enum EnrollmentFlow {
    private static let agentExecutablePath = ProboAgentHelperConstants.agentExecutablePath

    public static func runPreflight(rawURL: String) throws -> EnrollPreflightResult {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: agentExecutablePath)
        process.arguments = ["enroll-url", "--preflight", rawURL]

        let output = Pipe()
        process.standardOutput = output
        process.standardError = output

        try process.run()
        let data = try output.fileHandleForReading.readToEnd() ?? Data()
        process.waitUntilExit()

        guard process.terminationStatus == 0 else {
            let message = String(data: data, encoding: .utf8)?
                .trimmingCharacters(in: .whitespacesAndNewlines)
            throw HelperClientError.operationFailed(process.terminationStatus, message)
        }

        return try JSONDecoder().decode(EnrollPreflightResult.self, from: data)
    }

    public static func installViaHelper(preflight: EnrollPreflightResult) throws {
        try HelperClient.shared.install(
            serverURL: preflight.server,
            enrollmentToken: preflight.token,
            configDir: preflight.configDir
        )
    }
}
