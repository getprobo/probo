import AppKit
import SwiftUI

enum WorkspaceRegion: String, CaseIterable, Identifiable {
    case us
    case eu
    case selfHosted

    var id: String { rawValue }

    var manifestID: String {
        switch self {
        case .us:
            return "us"
        case .eu:
            return "eu"
        case .selfHosted:
            return "self_hosted"
        }
    }

    func definition(in manifest: RegionsManifest) -> RegionDefinition? {
        manifest.region(id: manifestID)
    }
}

struct EnrollmentPayload: Codable {
    let serverURL: String
    let enrollmentToken: String

    enum CodingKeys: String, CodingKey {
        case serverURL = "server_url"
        case enrollmentToken = "enrollment_token"
    }
}

struct RegionOptionCard: View {
    let definition: RegionDefinition
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Text(definition.flag)
                    .font(.system(size: 30))

                Text(definition.title)
                    .font(.subheadline.weight(.semibold))
                    .foregroundColor(.primary)
                    .multilineTextAlignment(.center)

                Text(definition.subtitle)
                    .font(.caption2)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .lineLimit(2)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 12)
            .padding(.horizontal, 8)
            .background(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .fill(
                        isSelected
                            ? Color.accentColor.opacity(0.10)
                            : Color(NSColor.controlBackgroundColor)
                    )
            )
            .overlay(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .stroke(
                        isSelected ? Color.accentColor : Color(NSColor.separatorColor),
                        lineWidth: isSelected ? 2 : 1
                    )
            )
        }
        .buttonStyle(PlainButtonStyle())
        .accessibilityLabel("\(definition.title), \(definition.subtitle)")
        .accessibilityAddTraits(isSelected ? [.isSelected] : [])
    }
}

struct EmployeePageLink: View {
    let hint: String
    let url: URL?
    let onOpen: (URL) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(hint)
                .font(.caption)
                .foregroundColor(.secondary)
                .fixedSize(horizontal: false, vertical: true)

            if let url {
                Button(action: { onOpen(url) }) {
                    HStack(spacing: 6) {
                        Image(systemName: "arrow.up.right.square")
                        Text("Open employee page")
                            .underline()
                    }
                    .font(.subheadline)
                }
                .buttonStyle(PlainButtonStyle())
                .foregroundColor(.accentColor)
            }
        }
        .padding(10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .fill(Color(NSColor.controlBackgroundColor))
        )
    }
}

struct EnrollmentView: View {
    let manifest: RegionsManifest

    @State private var region: WorkspaceRegion = .us
    @State private var customHost = ""
    @State private var token = ""
    @State private var errorMessage: String?

    let onComplete: (EnrollmentPayload?) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            VStack(alignment: .leading, spacing: 6) {
                Text("Connect this Mac to Probo")
                    .font(.title2.weight(.semibold))

                Text(
                    "Choose where your workspace is hosted, open your employee "
                        + "page to copy an enrollment token, then paste it below."
                )
                .font(.subheadline)
                .foregroundColor(.secondary)
                .fixedSize(horizontal: false, vertical: true)
            }

            VStack(alignment: .leading, spacing: 10) {
                Text("Workspace")
                    .font(.headline)

                HStack(spacing: 10) {
                    ForEach(WorkspaceRegion.allCases) { item in
                        if let definition = item.definition(in: manifest) {
                            RegionOptionCard(
                                definition: definition,
                                isSelected: region == item,
                                action: { region = item }
                            )
                        }
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityLabel("Workspace region")

                if region == .selfHosted {
                    TextField("probo.example.com", text: $customHost)
                        .textFieldStyle(RoundedBorderTextFieldStyle())
                }

                EmployeePageLink(
                    hint: employeeHint,
                    url: employeePageURL(for: region),
                    onOpen: { url in
                        NSWorkspace.shared.open(url)
                    }
                )
            }

            VStack(alignment: .leading, spacing: 8) {
                Text("Enrollment token")
                    .font(.headline)

                SecureField("Paste token here", text: $token)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
            }

            if let errorMessage {
                Text(errorMessage)
                    .font(.caption)
                    .foregroundColor(.red)
                    .fixedSize(horizontal: false, vertical: true)
            }

            HStack {
                Button("Cancel") {
                    onComplete(nil)
                }
                .keyboardShortcut(.cancelAction)

                Spacer()

                Button("Enroll device") {
                    submit()
                }
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(24)
        .frame(width: 500)
    }

    private var employeeHint: String {
        if employeePageURL(for: region) != nil {
            return manifest.employeePageHint
        }

        if region == .selfHosted {
            return manifest.selfHostedHostnameHint
        }

        return manifest.employeePageHint
    }

    private func submit() {
        errorMessage = nil

        guard let serverURL = resolveServerURL() else {
            return
        }

        let trimmedToken = token.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedToken.isEmpty else {
            errorMessage = "Enrollment token is required."
            return
        }

        onComplete(
            EnrollmentPayload(
                serverURL: serverURL,
                enrollmentToken: trimmedToken
            )
        )
    }

    private func employeePageURL(for region: WorkspaceRegion) -> URL? {
        guard let base = consoleBaseURL(for: region) else {
            return nil
        }

        return URL(string: base)
    }

    private func consoleBaseURL(for region: WorkspaceRegion) -> String? {
        switch region {
        case .us, .eu:
            return region.definition(in: manifest)?.serverURL
        case .selfHosted:
            return normalizeCustomHost(customHost)
        }
    }

    private func resolveServerURL() -> String? {
        switch region {
        case .us, .eu:
            return region.definition(in: manifest)?.serverURL
        case .selfHosted:
            let trimmed = customHost.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !trimmed.isEmpty else {
                errorMessage = "Enter your workspace hostname."
                return nil
            }

            if trimmed.contains("/") || trimmed.contains("?") || trimmed.contains("#") {
                errorMessage = "Hostname must not include a path or query string."
                return nil
            }

            guard let value = normalizeCustomHost(trimmed) else {
                errorMessage = "Enter a valid workspace hostname."
                return nil
            }

            return value
        }
    }

    private func normalizeCustomHost(_ host: String) -> String? {
        let trimmed = host.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            return nil
        }

        if trimmed.contains("/") || trimmed.contains("?") || trimmed.contains("#") {
            return nil
        }

        var value = trimmed
        if !value.lowercased().hasPrefix("http://") && !value.lowercased().hasPrefix("https://") {
            value = "https://\(value)"
        }

        if value.hasSuffix("/") {
            value.removeLast()
        }

        return value
    }
}
