import Darwin
import Foundation
import Security
import os

public enum AgentAuthError: LocalizedError {
    case missing
    case notRegularFile
    case missingRequirement
    case invalidRequirement
    case signatureRejected(OSStatus)

    public var errorDescription: String? {
        let path = ProboAgentHelperConstants.agentExecutablePath
        switch self {
        case .missing:
            return "probo-agent is not installed at \(path)"
        case .notRegularFile:
            return "\(path) is not a regular file"
        case .missingRequirement:
            return "missing agent code-signing requirement"
        case .invalidRequirement:
            return "invalid agent code-signing requirement"
        case .signatureRejected(let status):
            return "\(path) failed the Probo code-signing requirement (status=\(status))"
        }
    }
}

/// Checks the privileged agent binary before a helper or URL-handler exec.
public enum AgentAuth {
    private static let log = Logger(
        subsystem: "com.probo.agent.helper",
        category: "AgentAuth"
    )

    public static func verify() throws {
        try verify(path: ProboAgentHelperConstants.agentExecutablePath)
    }

    public static func verify(path: String) throws {
        try verifyPath(path: path)
        try verifySignature(path: path)
    }

    private static func verifyPath(path: String) throws {
        var fileStat = stat()
        guard lstat(path, &fileStat) == 0 else {
            log.error("reject agent binary (lstat failed)")
            throw AgentAuthError.missing
        }

        guard (fileStat.st_mode & S_IFMT) == S_IFREG else {
            log.error("reject agent binary (not a regular file)")
            throw AgentAuthError.notRegularFile
        }

        let parent = (path as NSString).deletingLastPathComponent
        var dirStat = stat()
        guard lstat(parent, &dirStat) == 0 else {
            log.error("reject agent binary (parent lstat failed)")
            throw AgentAuthError.missing
        }

        guard (dirStat.st_mode & S_IFMT) == S_IFDIR else {
            log.error("reject agent binary (parent is not a directory)")
            throw AgentAuthError.missing
        }
    }

    private static func verifySignature(path: String) throws {
        guard let requirement = agentRequirement() else {
            log.error("reject agent binary (missing requirement)")
            throw AgentAuthError.missingRequirement
        }

        var secRequirement: SecRequirement?
        guard
            SecRequirementCreateWithString(
                requirement as CFString,
                SecCSFlags(),
                &secRequirement
            ) == errSecSuccess, let secRequirement
        else {
            log.error("reject agent binary (invalid requirement string)")
            throw AgentAuthError.invalidRequirement
        }

        var staticCode: SecStaticCode?
        let url = URL(fileURLWithPath: path)
        let created = SecStaticCodeCreateWithPath(url as CFURL, SecCSFlags(), &staticCode)
        guard created == errSecSuccess, let staticCode else {
            log.error("reject agent binary (SecStaticCodeCreateWithPath=\(created))")
            throw AgentAuthError.signatureRejected(created)
        }

        let flags = SecCSFlags(rawValue: kSecCSCheckAllArchitectures)
        let check = SecStaticCodeCheckValidity(staticCode, flags, secRequirement)
        if check != errSecSuccess {
            log.error("reject agent binary (requirement not satisfied, status=\(check))")
            throw AgentAuthError.signatureRejected(check)
        }
    }

    private static func agentRequirement() -> String? {
        guard let teamID = ProboAgentSigningConstants.teamID, !teamID.isEmpty else {
            return nil
        }

        return """
            anchor apple generic and identifier "\(ProboAgentHelperConstants.agentIdentifier)" \
            and certificate leaf[subject.OU] = "\(teamID)"
            """
    }
}
