// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

namespace Probo.Agent.EnrollUI;

internal enum WorkspaceRegion
{
    Us,
    Eu,
    SelfHosted,
}

internal static class WorkspaceRegionExtensions
{
    internal static string ManifestID(this WorkspaceRegion region) => region switch
    {
        WorkspaceRegion.Us => "us",
        WorkspaceRegion.Eu => "eu",
        WorkspaceRegion.SelfHosted => "self_hosted",
        _ => throw new ArgumentOutOfRangeException(nameof(region), region, null),
    };

    internal static string? ConsoleBaseURL(
        this WorkspaceRegion region,
        RegionsManifest manifest,
        string customHost) => region switch
    {
        WorkspaceRegion.Us => manifest.Region(region).ServerURL,
        WorkspaceRegion.Eu => manifest.Region(region).ServerURL,
        WorkspaceRegion.SelfHosted => NormalizeCustomHost(customHost),
        _ => null,
    };

    internal static string? ResolveServerURL(
        this WorkspaceRegion region,
        RegionsManifest manifest,
        string customHost,
        out string? errorMessage)
    {
        errorMessage = null;

        switch (region)
        {
            case WorkspaceRegion.Us:
            case WorkspaceRegion.Eu:
                return manifest.Region(region).ServerURL;
            case WorkspaceRegion.SelfHosted:
                var trimmed = customHost.Trim();
                if (trimmed.Length == 0)
                {
                    errorMessage = "Enter your workspace hostname.";
                    return null;
                }

                if (trimmed.Contains('/') || trimmed.Contains('?') || trimmed.Contains('#'))
                {
                    errorMessage = "Hostname must not include a path or query string.";
                    return null;
                }

                var value = NormalizeCustomHost(trimmed);
                if (value is null)
                {
                    errorMessage = "Enter a valid workspace hostname.";
                    return null;
                }

                return value;
            default:
                return null;
        }
    }

    private static string? NormalizeCustomHost(string host)
    {
        var trimmed = host.Trim();
        if (trimmed.Length == 0)
        {
            return null;
        }

        if (trimmed.Contains('/') || trimmed.Contains('?') || trimmed.Contains('#'))
        {
            return null;
        }

        var value = trimmed;
        if (!value.StartsWith("http://", StringComparison.OrdinalIgnoreCase)
            && !value.StartsWith("https://", StringComparison.OrdinalIgnoreCase))
        {
            value = "https://" + value;
        }

        if (value.EndsWith('/'))
        {
            value = value[..^1];
        }

        return value;
    }
}
