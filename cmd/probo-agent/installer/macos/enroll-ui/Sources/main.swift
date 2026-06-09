import AppKit
import SwiftUI

final class AppDelegate: NSObject, NSApplicationDelegate, NSWindowDelegate {
    private var window: NSWindow?
    private var didFinish = false
    private let manifest: RegionsManifest

    init(manifest: RegionsManifest) {
        self.manifest = manifest
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApp.setActivationPolicy(.regular)

        let view = EnrollmentView(manifest: manifest) { payload in
            self.finish(payload: payload)
        }

        let hosting = NSHostingView(rootView: view)
        hosting.frame.size = hosting.fittingSize

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 500, height: hosting.fittingSize.height),
            styleMask: [.titled, .closable],
            backing: .buffered,
            defer: false
        )
        window.title = "Probo Device Posture Agent"
        window.contentView = hosting
        window.delegate = self
        window.center()
        window.makeKeyAndOrderFront(nil)
        window.isReleasedWhenClosed = false
        self.window = window

        NSApp.activate(ignoringOtherApps: true)
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        true
    }

    func windowWillClose(_ notification: Notification) {
        if notification.object as? NSWindow === window {
            finish(payload: nil)
        }
    }

    private func finish(payload: EnrollmentPayload?) {
        guard !didFinish else {
            return
        }

        didFinish = true

        guard let payload else {
            exit(1)
        }

        let encoder = JSONEncoder()
        encoder.outputFormatting = [.sortedKeys]

        guard let data = try? encoder.encode(payload),
              let json = String(data: data, encoding: .utf8)
        else {
            exit(2)
        }

        print(json)
        fflush(stdout)
        exit(0)
    }
}

do {
    let manifest = try RegionsManifest.load()
    let app = NSApplication.shared
    let delegate = AppDelegate(manifest: manifest)
    app.delegate = delegate
    app.run()
} catch {
    let alert = NSAlert()
    alert.messageText = "Probo Device Posture Agent"
    alert.informativeText = error.localizedDescription
    alert.alertStyle = .warning
    alert.runModal()
    exit(2)
}
