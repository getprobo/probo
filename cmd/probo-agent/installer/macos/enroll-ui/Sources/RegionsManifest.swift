import Foundation

struct RegionDefinition: Codable, Identifiable {
    let id: String
    let title: String
    let subtitle: String
    let flag: String
    let serverURL: String?

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case subtitle
        case flag
        case serverURL = "server_url"
    }
}

struct RegionsManifest: Codable {
    let regions: [RegionDefinition]
    let employeePageHint: String
    let selfHostedHostnameHint: String

    enum CodingKeys: String, CodingKey {
        case regions
        case employeePageHint = "employee_page_hint"
        case selfHostedHostnameHint = "self_hosted_hostname_hint"
    }

    func region(id: String) -> RegionDefinition? {
        regions.first { $0.id == id }
    }

    static func load() throws -> RegionsManifest {
        let executableURL = URL(fileURLWithPath: CommandLine.arguments[0])
        let installedURL = executableURL
            .deletingLastPathComponent()
            .appendingPathComponent("regions.json")

        if FileManager.default.fileExists(atPath: installedURL.path) {
            let data = try Data(contentsOf: installedURL)
            return try JSONDecoder().decode(RegionsManifest.self, from: data)
        }

        let sourceURL = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .appendingPathComponent("regions.json")

        if FileManager.default.fileExists(atPath: sourceURL.path) {
            let data = try Data(contentsOf: sourceURL)
            return try JSONDecoder().decode(RegionsManifest.self, from: data)
        }

        throw RegionsManifestError.notFound
    }
}

enum RegionsManifestError: LocalizedError {
    case notFound

    var errorDescription: String? {
        switch self {
        case .notFound:
            return "regions.json was not found next to the enrollment UI binary."
        }
    }
}
