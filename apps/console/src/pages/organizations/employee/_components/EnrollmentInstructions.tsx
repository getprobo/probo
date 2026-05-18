// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";

const SERVER_URL = "https://app.getprobo.com";
const RELEASE_BASE_URL
  = "https://github.com/getprobo/probo/releases/latest/download";

interface EnrollmentInstructionsProps {
  secret: string;
}

export function EnrollmentInstructions({ secret }: EnrollmentInstructionsProps) {
  const { __ } = useTranslate();

  const unixCommand = `# 1. Download and install the probo-agent binary
OS=$(uname -s); ARCH=$(uname -m | sed 's/aarch64/arm64/; s/amd64/x86_64/')
NAME="probo-agent_\${OS}_\${ARCH}"
curl -fsSL "${RELEASE_BASE_URL}/\${NAME}.tar.gz" -o /tmp/probo-agent.tar.gz
tar -xzf /tmp/probo-agent.tar.gz -C /tmp
sudo install -m 0755 "/tmp/\${NAME}/probo-agent" /usr/local/bin/probo-agent
rm -rf /tmp/probo-agent.tar.gz "/tmp/\${NAME}"

# 2. Enroll the device and start the agent service
sudo /usr/local/bin/probo-agent install \\
  --server ${SERVER_URL} \\
  --enrollment-token '${secret}'`;

  const windowsCommand = `# 1. Download and install the probo-agent binary
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'x86_64' }
$name = "probo-agent_Windows_$arch"
$zip  = "$env:TEMP\\probo-agent.zip"
$dst  = "$env:ProgramFiles\\Probo"
Invoke-WebRequest -Uri "${RELEASE_BASE_URL}/$name.zip" -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $env:TEMP -Force
New-Item -ItemType Directory -Force -Path $dst | Out-Null
Move-Item -Force "$env:TEMP\\$name\\probo-agent.exe" "$dst\\probo-agent.exe"
Remove-Item -Recurse -Force $zip, "$env:TEMP\\$name"

# 2. Enroll the device and start the agent service
& "$dst\\probo-agent.exe" install \`
  --server ${SERVER_URL} \`
  --enrollment-token '${secret}'`;

  return (
    <section className="border border-success-border bg-success-bg p-4 rounded">
      <h2 className="font-medium mb-2">{__("Enrollment token generated")}</h2>
      <p className="text-sm text-tertiary mb-2">
        {__(
          "Copy this token now. It is shown only once — even Probo cannot recover it.",
        )}
      </p>
      <pre className="text-xs bg-surface-default p-3 rounded break-all">
        {secret}
      </pre>

      <h3 className="mt-4 mb-1 text-sm font-medium">
        {__("Install on macOS or Linux (run from a shell with sudo access)")}
      </h3>
      <pre className="text-xs bg-surface-default p-3 rounded whitespace-pre overflow-x-auto">
        {unixCommand}
      </pre>

      <h3 className="mt-4 mb-1 text-sm font-medium">
        {__("Install on Windows (run from an elevated PowerShell session)")}
      </h3>
      <pre className="text-xs bg-surface-default p-3 rounded whitespace-pre overflow-x-auto">
        {windowsCommand}
      </pre>

      <p className="text-xs text-tertiary mt-3">
        {__(
          "The token is passed as a CLI flag (not via curl-piped-to-shell or sudo env vars). Once installed, the agent self-updates from GitHub Releases with cosign signature verification.",
        )}
      </p>
    </section>
  );
}
