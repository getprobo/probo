import AppKit
import Foundation
import HelperClient

private final class URLHandlerApp: NSObject, NSApplicationDelegate {
    private var didReceiveURL = false
    private var idleTimer: Timer?

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
        NSLog("probo-agent url-handler: launched")
        idleTimer = Timer.scheduledTimer(withTimeInterval: 15, repeats: false) { [weak self] _ in
            guard let self, !self.didReceiveURL else { return }
            NSLog("probo-agent url-handler: no URL received; exiting")
            NSApp.terminate(nil)
        }
    }

    func application(_ application: NSApplication, open urls: [URL]) {
        guard let rawURL = urls.first?.absoluteString else { return }
        beginEnrollmentIfNeeded(rawURL: rawURL)
    }

    @objc private func handleGetURLEvent(
        _ event: NSAppleEventDescriptor,
        withReplyEvent replyEvent: NSAppleEventDescriptor
    ) {
        guard let rawURL = event.paramDescriptor(forKeyword: keyDirectObject)?.stringValue else {
            presentFailure("Enrollment link is missing.")
            return
        }

        beginEnrollmentIfNeeded(rawURL: rawURL)
    }

    private func beginEnrollmentIfNeeded(rawURL: String) {
        guard !didReceiveURL else { return }
        didReceiveURL = true
        idleTimer?.invalidate()
        idleTimer = nil
        runEnrollment(for: rawURL)
    }

    private func runEnrollment(for rawURL: String) {
        NSLog("probo-agent url-handler: starting enrollment")

        DispatchQueue.global(qos: .userInitiated).async {
            do {
                NSLog("probo-agent url-handler: preflight…")
                let preflight = try EnrollmentFlow.runPreflight(rawURL: rawURL)
                if preflight.alreadyEnrolled {
                    NSLog("probo-agent url-handler: already enrolled")
                    DispatchQueue.main.async {
                        NSApp.terminate(nil)
                    }
                    return
                }

                NSLog("probo-agent url-handler: install via helper…")
                try EnrollmentFlow.installViaHelper(preflight: preflight)
                NSLog("probo-agent url-handler: install completed")

                DispatchQueue.main.async {
                    NSApp.terminate(nil)
                }
            } catch {
                NSLog(
                    "probo-agent url-handler: enrollment failed: %@",
                    error.localizedDescription
                )
                DispatchQueue.main.async {
                    self.presentFailure(error.localizedDescription)
                }
            }
        }
    }

    private func presentFailure(_ message: String) {
        presentAlert(title: "Enrollment failed", message: message, style: .warning)
    }

    private func presentAlert(title: String, message: String, style: NSAlert.Style) {
        // LSUIElement / .accessory apps otherwise show alerts that never appear.
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)

        let alert = NSAlert()
        alert.messageText = title
        alert.informativeText = message
        alert.alertStyle = style
        alert.runModal()

        NSApp.setActivationPolicy(.accessory)
        NSApp.terminate(nil)
    }
}

let app = NSApplication.shared
private let delegate = URLHandlerApp()
app.delegate = delegate
app.setActivationPolicy(.accessory)
app.run()
