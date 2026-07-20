import Foundation
import ProboAgentShared

public enum HelperClientError: LocalizedError {
    case helperNotInstalled
    case connectionFailed(String)
    case operationFailed(Int32, String?)

    public var errorDescription: String? {
        switch self {
        case .helperNotInstalled:
            return """
                Privileged helper is not installed. Reinstall the Probo Agent \
                package, then try enrollment again.
                """
        case .connectionFailed(let message):
            return "Cannot connect to privileged helper: \(message)"
        case .operationFailed(let code, let message):
            if let message, !message.isEmpty {
                return message
            }
            return "Privileged operation failed (exit \(code))."
        }
    }
}

/// Once-only sync bridge for an in-flight XPC call. Reply and proxy error
/// handlers both finish here so disconnects unblock waiters immediately.
private final class XPCCallCompletion: @unchecked Sendable {
    private let lock = NSLock()
    private let semaphore = DispatchSemaphore(value: 0)
    private var finished = false
    private var error: Error?

    func succeed() {
        complete(nil)
    }

    func fail(_ error: Error) {
        complete(error)
    }

    private func complete(_ error: Error?) {
        lock.lock()
        defer { lock.unlock() }
        guard !finished else { return }
        finished = true
        self.error = error
        semaphore.signal()
    }

    /// Waits for succeed/fail. Throws the stored error, or `timedOut` on timeout.
    func wait(timeout: TimeInterval, timedOut: @autoclosure () -> Error) throws {
        if semaphore.wait(timeout: .now() + timeout) == .timedOut {
            throw timedOut()
        }
        if let error {
            throw error
        }
    }

    /// Waits for succeed/fail. Returns `false` on timeout; throws the stored error.
    func wait(timeout: TimeInterval) throws -> Bool {
        if semaphore.wait(timeout: .now() + timeout) == .timedOut {
            return false
        }
        if let error {
            throw error
        }
        return true
    }
}

final public class HelperClient {
    public static let shared = HelperClient()

    /// Serializes public install within this process so privileged helper
    /// work never overlaps. Does not coordinate across processes.
    private let operationLock = NSLock()

    private init() {}

    public func install(
        serverURL: String,
        enrollmentToken: String,
        configDir: String = ProboAgentHelperConstants.defaultConfigDir
    ) throws {
        operationLock.lock()
        defer { operationLock.unlock() }

        try ensureHelperReady()
        try withRemoteProxy { proxy, completion in
            proxy.install(
                serverURL: serverURL,
                enrollmentToken: enrollmentToken,
                configDir: configDir
            ) { exitCode, output in
                if exitCode != 0 {
                    completion.fail(HelperClientError.operationFailed(exitCode, output))
                } else {
                    completion.succeed()
                }
            }

            // Headroom over probo-agent install's 60s deadline plus local
            // service/tray setup so we report the command's real outcome.
            try completion.wait(
                timeout: 120,
                timedOut: HelperClientError.connectionFailed(
                    "install timed out waiting for privileged helper"
                ))
        }
    }

    /// The helper is installed by the PKG postinstall (as root). Enrollment
    /// never calls SMJobBless — no admin prompt on the browser path.
    private func ensureHelperReady() throws {
        guard isHelperInstalled() else {
            throw HelperClientError.helperNotInstalled
        }

        let installedVersion = try installedHelperVersion()
        if installedVersion == nil {
            throw HelperClientError.connectionFailed(
                "helper is installed but not responding; reinstall the Probo Agent package"
            )
        }

        if installedVersion != ProboAgentHelperConstants.helperVersion {
            NSLog(
                "probo-agent helper client: version mismatch (installed=%@ expected=%@)",
                installedVersion ?? "nil",
                ProboAgentHelperConstants.helperVersion
            )
        }

        try verifyHelperResponds()
    }

    private func isHelperInstalled() -> Bool {
        FileManager.default.fileExists(
            atPath: "/Library/PrivilegedHelperTools/\(ProboAgentHelperConstants.helperLabel)"
        )
    }

    private func installedHelperVersion() throws -> String? {
        try withRemoteProxy { proxy, completion in
            var version: String?

            proxy.getVersion { value in
                version = value
                completion.succeed()
            }

            guard try completion.wait(timeout: 10) else {
                return nil
            }

            return version
        }
    }

    private func verifyHelperResponds() throws {
        try withRemoteProxy { proxy, completion in
            var ok = false

            proxy.ping { value in
                ok = value
                completion.succeed()
            }

            try completion.wait(
                timeout: 10,
                timedOut: HelperClientError.connectionFailed(
                    "helper did not respond to ping (timeout)")
            )
            if !ok {
                throw HelperClientError.connectionFailed("helper did not respond to ping")
            }
        }
    }

    /// Creates a dedicated XPC connection for the duration of `body`, then
    /// invalidates it. Callers must finish waiting for replies inside `body`
    /// so the connection outlives the reply.
    private func withRemoteProxy<T>(
        _ body: (ProboAgentHelperProtocol, XPCCallCompletion) throws -> T
    ) throws -> T {
        let connection = NSXPCConnection(
            machServiceName: ProboAgentHelperConstants.machServiceName,
            options: .privileged
        )
        connection.remoteObjectInterface = NSXPCInterface(with: ProboAgentHelperProtocol.self)
        connection.resume()
        defer { connection.invalidate() }

        let completion = XPCCallCompletion()
        guard
            let proxy = connection.remoteObjectProxyWithErrorHandler({ error in
                NSLog(
                    "probo-agent helper XPC error: %@",
                    error.localizedDescription
                )
                completion.fail(
                    HelperClientError.connectionFailed(error.localizedDescription)
                )
            }) as? ProboAgentHelperProtocol
        else {
            throw HelperClientError.connectionFailed("cannot create remote proxy")
        }

        return try body(proxy, completion)
    }
}
