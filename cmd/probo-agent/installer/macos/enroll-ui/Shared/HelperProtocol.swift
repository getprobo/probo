import Foundation

@objc(ProboAgentHelperProtocol)
public protocol ProboAgentHelperProtocol: NSObjectProtocol {
    func getVersion(withReply reply: @escaping (String) -> Void)
    func ping(withReply reply: @escaping (Bool) -> Void)
    func install(
        serverURL: String,
        enrollmentToken: String,
        configDir: String,
        withReply reply: @escaping (Int32, String?) -> Void
    )
}
