import Foundation

public enum ProboAgentHelperConstants {
    public static let machServiceName = "com.probo.agent.helper"
    public static let helperLabel = "com.probo.agent.helper"
    public static let clientBundleID = "com.probo.agent.url-handler"
    public static let agentExecutablePath = "/usr/local/bin/probo-agent"
    public static let defaultConfigDir = "/var/lib/probo-agent"
    public static let enrolledMarkerPath = "/var/run/probo-agent/enrolled"

    public static let helperVersion = ProboAgentHelperVersion.value
}
