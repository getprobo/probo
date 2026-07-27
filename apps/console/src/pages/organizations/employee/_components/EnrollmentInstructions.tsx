// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Button, Card, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";

const RELEASE_BASE_URL
  = "https://github.com/getprobo/probo/releases/latest/download";

interface EnrollmentInstructionsProps {
  enrollmentToken: string;
  serverUrl: string;
}

export function EnrollmentInstructions(
  { enrollmentToken, serverUrl }: EnrollmentInstructionsProps,
) {
  const { t } = useTranslation();

  const unixCommand = `# 1. Download and install the probo-agent binary
OS=$(uname -s); ARCH=$(uname -m | sed 's/aarch64/arm64/; s/amd64/x86_64/')
NAME="probo-agent_\${OS}_\${ARCH}"
curl -fsSL "${RELEASE_BASE_URL}/\${NAME}.tar.gz" -o /tmp/probo-agent.tar.gz
tar -xzf /tmp/probo-agent.tar.gz -C /tmp
sudo install -m 0755 "/tmp/\${NAME}/probo-agent" /usr/local/bin/probo-agent
rm -rf /tmp/probo-agent.tar.gz "/tmp/\${NAME}"

# 2. Configure the device and start the agent service
sudo /usr/local/bin/probo-agent install \\
  --server ${serverUrl} \\
  --enrollment-token '${enrollmentToken}'`;

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

# 2. Configure the device and start the agent service
& "$dst\\probo-agent.exe" install \`
  --server ${serverUrl} \`
  --enrollment-token '${enrollmentToken}'`;

  return (
    <div className="space-y-4">
      <h2 className="font-medium">{t("deviceEnrollment.token.title")}</h2>
      <p className="text-sm text-txt-secondary">
        {t("deviceEnrollment.token.description")}
      </p>
      <CopyableCodeBlock code={enrollmentToken} />

      <details>
        <summary className="cursor-pointer text-sm font-medium">
          {t("deviceEnrollment.token.manualInstall")}
        </summary>

        <div className="mt-4 space-y-4">
          <div className="space-y-2">
            <h3 className="text-sm font-medium">
              {t("deviceEnrollment.token.installUnix")}
            </h3>
            <CopyableCodeBlock code={unixCommand} />
          </div>

          <div className="space-y-2">
            <h3 className="text-sm font-medium">
              {t("deviceEnrollment.token.installWindows")}
            </h3>
            <CopyableCodeBlock code={windowsCommand} />
          </div>

          <p className="text-xs text-txt-secondary">
            {t("deviceEnrollment.token.securityNotice")}
          </p>
        </div>
      </details>
    </div>
  );
}

function CopyableCodeBlock({ code }: { code: string }) {
  const { t } = useTranslation();
  const { toast } = useToast();

  const handleCopy = () => {
    navigator.clipboard.writeText(code).then(
      () => {
        toast({
          title: t("common.messages.copied"),
          description: t("common.messages.copiedToClipboard"),
          variant: "success",
        });
      },
      () => {
        toast({
          title: t("common.error"),
          description: t("deviceEnrollment.errors.copyToClipboard"),
          variant: "error",
        });
      },
    );
  };

  return (
    <Card className="rounded-lg border">
      <div className="flex items-center justify-end border-b border-border-low px-1 py-1">
        <Button type="button" variant="secondary" onClick={handleCopy}>
          {t("common.actions.copy")}
        </Button>
      </div>
      <pre className="overflow-x-auto whitespace-pre p-4 text-sm font-mono rounded-b-lg text-invert bg-accent">
        <code>{code}</code>
      </pre>
    </Card>
  );
}
