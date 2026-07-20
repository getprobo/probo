import Foundation
import ProboAgentShared
import os

private let log = Logger(subsystem: "com.probo.agent.helper", category: "main")

let helper = Helper()
let listener = NSXPCListener(machServiceName: ProboAgentHelperConstants.machServiceName)
listener.delegate = helper
listener.resume()
log.info("listening on \(ProboAgentHelperConstants.machServiceName, privacy: .public)")

RunLoop.main.run()
