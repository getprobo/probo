import AppKit
import Darwin
import Foundation

// Fixed install location written by the macOS PKG postinstall script
// (cmd/probo-agent/installer/macos/scripts/postinstall, BINARY).
private let agentExecutablePath = "/usr/local/bin/probo-agent"

private enum EnrollmentCallbackState: String, Codable {
    case success
    case failure
}

private struct EnrollmentCallbackPayload: Codable {
    let state: EnrollmentCallbackState
    let message: String?
}

private enum EnrollmentCallbackStore {
    static var statusFileURL: URL {
        FileManager.default.temporaryDirectory
            .appendingPathComponent("probo-agent-enrollment-status.json")
    }

    static var lockFileURL: URL {
        FileManager.default.temporaryDirectory
            .appendingPathComponent("probo-agent-enrollment-ui.lock")
    }

    static func isWizardRunning() -> Bool {
        guard
            let data = try? Data(contentsOf: lockFileURL),
            let pidText = String(data: data, encoding: .utf8)?
                .trimmingCharacters(in: .whitespacesAndNewlines),
            let pid = Int32(pidText),
            pid > 0
        else {
            return false
        }

        return kill(pid, 0) == 0
    }

    static func writeStatus(
        state: EnrollmentCallbackState,
        message: String?
    ) {
        let payload = EnrollmentCallbackPayload(state: state, message: message)
        guard let data = try? JSONEncoder().encode(payload) else {
            return
        }

        try? data.write(to: statusFileURL, options: [.atomic])
    }
}

private final class URLHandlerApp: NSObject, NSApplicationDelegate {
    private var didReceiveURL = false

    override init() {
        super.init()

        NSAppleEventManager.shared().setEventHandler(
            self,
            andSelector: #selector(handleGetURLEvent(_:withReplyEvent:)),
            forEventClass: AEEventClass(kInternetEventClass),
            andEventID: AEEventID(kAEGetURL)
        )
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        Timer.scheduledTimer(withTimeInterval: 10, repeats: false) { _ in
            if !self.didReceiveURL {
                NSApp.terminate(nil)
            }
        }
    }

    @objc private func handleGetURLEvent(
        _ event: NSAppleEventDescriptor,
        withReplyEvent replyEvent: NSAppleEventDescriptor
    ) {
        guard let rawURL = event.paramDescriptor(forKeyword: keyDirectObject)?.stringValue else {
            reportFailure("Enrollment link is missing.")
            return
        }

        didReceiveURL = true
        runEnrollment(for: rawURL)
    }

    private func runEnrollment(for rawURL: String) {
        let shouldNotifyWizard = EnrollmentCallbackStore.isWizardRunning()

        DispatchQueue.global(qos: .userInitiated).async {
            let process = Process()
            process.executableURL = URL(fileURLWithPath: agentExecutablePath)
            process.arguments = ["enroll-url", rawURL]

            let output = Pipe()
            process.standardOutput = output
            process.standardError = output

            do {
                try process.run()
                process.waitUntilExit()
            } catch {
                DispatchQueue.main.async {
                    self.reportFailure(
                        self.sanitizedFailureMessage(error.localizedDescription),
                        shouldNotifyWizard: shouldNotifyWizard
                    )
                }
                return
            }

            guard process.terminationStatus == 0 else {
                let data = output.fileHandleForReading.readDataToEndOfFile()
                let message = String(data: data, encoding: .utf8)?
                    .trimmingCharacters(in: .whitespacesAndNewlines)

                DispatchQueue.main.async {
                    self.reportFailure(
                        self.sanitizedFailureMessage(message),
                        shouldNotifyWizard: shouldNotifyWizard
                    )
                }
                return
            }

            DispatchQueue.main.async {
                if shouldNotifyWizard {
                    EnrollmentCallbackStore.writeStatus(state: .success, message: nil)
                }
                NSApp.terminate(nil)
            }
        }
    }

    private func reportFailure(_ message: String, shouldNotifyWizard: Bool = EnrollmentCallbackStore.isWizardRunning()) {
        if shouldNotifyWizard {
            EnrollmentCallbackStore.writeStatus(state: .failure, message: message)
            NSApp.terminate(nil)
            return
        }

        showError(message)
    }

    private func sanitizedFailureMessage(_ raw: String?) -> String {
        guard let raw else {
            return "Enrollment failed. Please try again."
        }

        let normalized = raw.lowercased()
        if normalized.contains("already enrolled") {
            return "This device is already enrolled."
        }
        if normalized.contains("token") && normalized.contains("missing") {
            return "Enrollment token is missing or invalid."
        }

        return "Enrollment failed. Please try again."
    }

    private func showError(_ message: String) {
        let alert = NSAlert()
        alert.messageText = "Enrollment failed"
        alert.informativeText = message
        alert.alertStyle = .warning
        NSApp.activate(ignoringOtherApps: true)
        alert.runModal()
        NSApp.terminate(nil)
    }
}

let app = NSApplication.shared
private let delegate = URLHandlerApp()
app.delegate = delegate
app.setActivationPolicy(.accessory)
app.run()
