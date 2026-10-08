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

import { WarningCircleIcon } from "@phosphor-icons/react";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { SourceConnectionIssue_connector$key } from "#/__generated__/core/SourceConnectionIssue_connector.graphql";
import type { ConnectionIssueKey } from "#/pages/organizations/_lib/connectorStatus";
import { ConnectorDocumentationLink } from "#/pages/organizations/settings/integrations/_components/ConnectorDocumentationLink";

import { connectionIssue } from "../sources/_components/variants";

const connectionIssueFragment = graphql`
  fragment SourceConnectionIssue_connector on Connector {
    displayName
    ...ConnectorDocumentationLink_connector
  }
`;

interface SourceConnectionIssueProps {
  connectorKey: SourceConnectionIssue_connector$key;
  issueKey: ConnectionIssueKey;
  reconnectUrl: string | null;
}

export function SourceConnectionIssue({
  connectorKey,
  issueKey,
  reconnectUrl,
}: SourceConnectionIssueProps) {
  const { t } = useTranslation();
  const connector = useFragment(connectionIssueFragment, connectorKey);
  const provider = connector.displayName;
  const { root, body, copy } = connectionIssue();

  const unavailable = "accessReviewSourceRow.organizations.unavailable";

  return (
    <Callout
      className={root()}
      size={1}
      variant="soft"
      color="red"
      icon={<WarningCircleIcon weight="fill" />}
    >
      <div className={body()}>
        <div className={copy()}>
          <Heading level={3} size={2} weight="medium" color="current">
            {t(`${unavailable}.${issueKey}Title`, { provider })}
          </Heading>
          <Text size={1} color="current">
            {t(`${unavailable}.${issueKey}Description`, { provider })}
          </Text>
          {issueKey !== "reconnect" && (
            <ConnectorDocumentationLink connectorKey={connector} />
          )}
        </div>
        {reconnectUrl && (
          <ButtonAnchor href={reconnectUrl} size={1}>
            {t("accessReviewSourceRow.actions.reconnect")}
          </ButtonAnchor>
        )}
      </div>
    </Callout>
  );
}
