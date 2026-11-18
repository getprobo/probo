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

import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import type { ConnectorDeleteDialogSourcesQuery } from "#/__generated__/core/ConnectorDeleteDialogSourcesQuery.graphql";
import { accessReviewSourceLabel } from "#/pages/organizations/_lib/sourceLabel";

export const connectorDeleteDialogSourcesQuery = graphql`
  query ConnectorDeleteDialogSourcesQuery($connectorId: ID!) @throwOnFieldError {
    node(id: $connectorId) {
      ... on Connector {
        accessReviewSources {
          id
          name
          connectorAccount {
            name
          }
        }
      }
    }
  }
`;

interface ConnectorDeleteDialogSourcesProps {
  queryRef: PreloadedQuery<ConnectorDeleteDialogSourcesQuery>;
}

export function ConnectorDeleteDialogSources({
  queryRef,
}: ConnectorDeleteDialogSourcesProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { node } = usePreloadedQuery<ConnectorDeleteDialogSourcesQuery>(
    connectorDeleteDialogSourcesQuery,
    queryRef,
  );
  const sources = node.accessReviewSources ?? [];

  if (sources.length === 0) {
    return null;
  }

  return (
    <DialogBody className="flex flex-col gap-3">
      <Text size={2}>
        {t("detailsPage.delete.sources", { count: sources.length })}
      </Text>
      <List>
        {sources.map(source => (
          <ListItem key={source.id}>
            <ListItemContent>
              <Text size={2}>
                {accessReviewSourceLabel({
                  name: source.name,
                  accountName: source.connectorAccount?.name,
                })}
              </Text>
            </ListItemContent>
          </ListItem>
        ))}
      </List>
    </DialogBody>
  );
}
