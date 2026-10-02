// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { CopyIcon } from "@phosphor-icons/react";
import { IconWarning, useToast } from "@probo/ui";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { ConnectionIssueKey } from "#/pages/organizations/_lib/connectorStatus";

import { connectorCard } from "../variants";

interface ConnectorProbeErrorProps {
  issues: readonly ConnectionIssueKey[];
  provider: string;
}

export function ConnectorProbeError({
  issues,
  provider,
}: ConnectorProbeErrorProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const { probeError, probeErrorHeader, probeErrorRow, probeErrorText } = connectorCard();

  async function copyError(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      return;
    }
    toast({
      title: t("detailsPage.probeError.copied"),
      description: t("detailsPage.probeError.copied"),
      variant: "success",
    });
  }

  return (
    <div className={probeError()}>
      <div className={probeErrorHeader()}>
        <span className="text-red-11" aria-hidden>
          <IconWarning size={16} className="shrink-0" />
        </span>
      </div>
      {issues.map((issue) => {
        const message = t(`listPage.connectionIssues.${issue}`, { provider });
        return (
          <div key={issue} className={probeErrorRow()}>
            <Text size={2} color="neutral" className={probeErrorText()}>
              {message}
            </Text>
            <IconButton
              size={1}
              variant="soft"
              color="neutral"
              aria-label={t("detailsPage.probeError.copy")}
              onClick={() => void copyError(message)}
            >
              <CopyIcon />
            </IconButton>
          </div>
        );
      })}
    </div>
  );
}
