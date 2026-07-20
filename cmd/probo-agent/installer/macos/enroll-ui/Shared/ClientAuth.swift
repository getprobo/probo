import Foundation
import Security
import os

public enum ClientAuth {
    private static let log = Logger(
        subsystem: "com.probo.agent.helper",
        category: "ClientAuth"
    )

    /// Validates an incoming XPC connection from the Probo Agent URL handler app.
    public static func accepts(connection: NSXPCConnection) -> Bool {
        guard let code = copyGuestCode(for: connection) else {
            return false
        }

        // Developer ID + hardened runtime sets CS_RUNTIME on the code directory
        // (kSecCodeInfoFlags). kSecCodeInfoStatus is only present for some
        // dynamic queries and was missing here, which rejected every client.
        if !hasAcceptableCodeStatus(code) {
            return false
        }

        guard let requirement = clientRequirement() else {
            log.error("reject XPC client (missing client requirement)")
            return false
        }

        var secRequirement: SecRequirement?
        guard
            SecRequirementCreateWithString(
                requirement as CFString,
                SecCSFlags(),
                &secRequirement
            ) == errSecSuccess, let secRequirement
        else {
            log.error("reject XPC client (invalid requirement string)")
            return false
        }

        let check = SecCodeCheckValidity(code, SecCSFlags(), secRequirement)
        if check != errSecSuccess {
            log.error("reject XPC client (requirement not satisfied, status=\(check))")
            return false
        }

        log.info("accepted XPC client pid=\(connection.processIdentifier, privacy: .public)")
        return true
    }

    private static func copyGuestCode(for connection: NSXPCConnection) -> SecCode? {
        if let token = auditToken(from: connection) {
            var mutableToken = token
            let tokenData = withUnsafeBytes(of: &mutableToken) { Data($0) }
            if let code = copyGuestCode(attributes: [
                kSecGuestAttributeAudit as String: tokenData
            ]) {
                return code
            }
            log.info("audit-token guest lookup failed; trying pid")
        } else {
            log.info("no audit token on XPC connection; trying pid")
        }

        let pid = connection.processIdentifier
        guard pid > 0 else {
            log.error("reject XPC client (invalid pid)")
            return nil
        }

        return copyGuestCode(attributes: [
            kSecGuestAttributePid as String: NSNumber(value: pid)
        ])
    }

    private static func copyGuestCode(attributes: [String: Any]) -> SecCode? {
        var code: SecCode?
        let status = SecCodeCopyGuestWithAttributes(
            nil,
            attributes as CFDictionary,
            SecCSFlags(),
            &code
        )
        guard status == errSecSuccess else {
            log.error("reject XPC client (SecCodeCopyGuestWithAttributes=\(status))")
            return nil
        }
        return code
    }

    private static func hasAcceptableCodeStatus(_ code: SecCode) -> Bool {
        var staticCode: SecStaticCode?
        guard SecCodeCopyStaticCode(code, SecCSFlags(), &staticCode) == errSecSuccess,
            let staticCode
        else {
            log.error("reject XPC client (cannot copy static code)")
            return false
        }

        var csInfo: CFDictionary?
        guard
            SecCodeCopySigningInformation(
                staticCode,
                SecCSFlags(rawValue: kSecCSDynamicInformation),
                &csInfo
            ) == errSecSuccess,
            let info = csInfo as? [String: Any]
        else {
            log.error("reject XPC client (cannot read signing information)")
            return false
        }

        // Prefer dynamic status when present; fall back to code-directory flags
        // (where hardened-runtime CS_RUNTIME lives for Developer ID binaries).
        let statusValue = uint32Value(info[kSecCodeInfoStatus as String])
        let flagsValue = uint32Value(info[kSecCodeInfoFlags as String])
        let bits = statusValue ?? flagsValue

        guard let bits else {
            log.error("reject XPC client (no status/flags in signing info)")
            return false
        }

        // Accept hardened-runtime clients (Developer ID + --options runtime sets
        // CS_RUNTIME). Also accept the older CS_HARD|CS_KILL pair.
        let csHard: UInt32 = 0x100
        let csKill: UInt32 = 0x200
        let csRuntime: UInt32 = 0x10000
        let hasHardKill = (bits & (csHard | csKill)) == (csHard | csKill)
        let hasRuntime = (bits & csRuntime) == csRuntime
        if !hasHardKill && !hasRuntime {
            log.error(
                "reject XPC client (bits=0x\(String(bits, radix: 16)), need runtime or hard|kill)"
            )
            return false
        }

        return true
    }

    private static func uint32Value(_ value: Any?) -> UInt32? {
        switch value {
        case let number as NSNumber:
            return number.uint32Value
        case let value as UInt32:
            return value
        case let value as Int:
            return UInt32(truncatingIfNeeded: value)
        default:
            return nil
        }
    }

    /// Reads NSXPCConnection.auditToken across SDK/runtime differences.
    /// Modern macOS exposes it as an ObjC property; KVC may return Data or NSValue.
    private static func auditToken(from connection: NSXPCConnection) -> audit_token_t? {
        if let data = connection.value(forKey: "auditToken") as? Data,
            data.count == MemoryLayout<audit_token_t>.size
        {
            return data.withUnsafeBytes { raw in
                raw.load(as: audit_token_t.self)
            }
        }

        if let value = connection.value(forKey: "auditToken") as? NSValue {
            var token = audit_token_t()
            value.getValue(&token)
            return token
        }

        return nil
    }

    private static func clientRequirement() -> String? {
        guard let teamID = ProboAgentSigningConstants.teamID, !teamID.isEmpty else {
            return nil
        }
        return """
            anchor apple generic and identifier "\(ProboAgentHelperConstants.clientBundleID)" \
            and certificate leaf[subject.OU] = "\(teamID)"
            """
    }
}
