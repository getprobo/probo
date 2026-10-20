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

import { ErrorNotice } from "@probo/ui/src/v2/ErrorNotice/ErrorNotice";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorProbeError_connector$key } from "#/__generated__/core/ConnectorProbeError_connector.graphql";
import type { ConnectionIssueKey } from "#/pages/organizations/_lib/connectorStatus";

const connectorProbeErrorFragment = graphql`
  fragment ConnectorProbeError_connector on Connector {
    displayName
  }
`;

interface ConnectorProbeErrorProps {
  connectorKey: ConnectorProbeError_connector$key;
  issues: readonly ConnectionIssueKey[];
}

export function ConnectorProbeError({
  connectorKey,
  issues,
}: ConnectorProbeErrorProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const connector = useFragment(connectorProbeErrorFragment, connectorKey);
  const toast = useToast();
  const provider = connector.displayName;

  return (
    <ErrorNotice
      className="pointer-events-auto relative z-1"
      messages={issues.map(issue => t(`listPage.connectionIssues.${issue}`, { provider }))}
      copyLabel={t("detailsPage.probeError.copy")}
      onCopied={() => {
        toast.add({
          title: t("detailsPage.probeError.copied"),
          type: "success",
        });
      }}
    />
  );
}
