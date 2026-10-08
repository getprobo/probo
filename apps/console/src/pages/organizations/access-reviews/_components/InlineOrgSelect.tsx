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

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";

import type { InlineOrgSelect_source$key } from "#/__generated__/core/InlineOrgSelect_source.graphql";
import type { InlineOrgSelectQuery } from "#/__generated__/core/InlineOrgSelectQuery.graphql";
import type { SourceConnectionIssue_connector$key } from "#/__generated__/core/SourceConnectionIssue_connector.graphql";
import type { ConnectorConnectionStatus } from "#/pages/organizations/_lib/connectorStatus";

import { connectionIssue } from "../_lib/connectionIssue";
import { sourceListItem } from "../sources/_components/variants";

import { CapturedOrganization } from "./CapturedOrganization";
import { ProviderOrganizationsEmpty } from "./ProviderOrganizationsEmpty";
import { SourceConnectionIssue } from "./SourceConnectionIssue";

const organizationsFragment = graphql`
  fragment InlineOrgSelect_source on AccessReviewSource {
    selectedOrganization
    providerOrganizations {
      status
      nodes {
        slug
        displayName
      }
    }
    ...CapturedOrganization_source
    ...ProviderOrganizationsEmpty_source
  }
`;

export const inlineOrgSelectQuery = graphql`
  query InlineOrgSelectQuery($accessReviewSourceId: ID!) {
    node(id: $accessReviewSourceId) @required(action: THROW) {
      ... on AccessReviewSource {
        ...InlineOrgSelect_source
      }
    }
  }
`;

interface InlineOrgSelectProps {
  queryRef: PreloadedQuery<InlineOrgSelectQuery>;
  connectionStatus: ConnectorConnectionStatus;
  canReconnect: boolean;
  reconnectUrl: string | null;
  connectorKey: SourceConnectionIssue_connector$key;
  providerName: string;
  onSelect: (slug: string) => void;
}

export function InlineOrgSelect({
  queryRef,
  connectionStatus,
  canReconnect,
  reconnectUrl,
  connectorKey,
  providerName,
  onSelect,
}: InlineOrgSelectProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<InlineOrgSelectQuery>(inlineOrgSelectQuery, queryRef);

  const source
    = useFragment<InlineOrgSelect_source$key>(
      organizationsFragment,
      data.node,
    );
  const providerOrganizations = source.providerOrganizations;

  switch (providerOrganizations?.status) {
    case "AVAILABLE": {
      const { organizationSelect } = sourceListItem();
      const options = [...providerOrganizations.nodes];
      if (
        source.selectedOrganization
        && !options.some(option => option.slug === source.selectedOrganization)
      ) {
        options.unshift({
          slug: source.selectedOrganization,
          displayName: source.selectedOrganization,
        });
      }

      return (
        <div className={organizationSelect()}>
          <Select
            value={source.selectedOrganization}
            onValueChange={(value) => {
              if (value != null) {
                onSelect(value);
              }
            }}
          >
            <SelectTrigger size={1} aria-label={t("accessReviewSourceRow.selectOrganization")}>
              {(value: string | null) => {
                if (value == null) {
                  return (
                    <Text size={1} color="faint">
                      {t("accessReviewSourceRow.selectOrganization")}
                    </Text>
                  );
                }

                return options.find(option => option.slug === value)?.displayName ?? value;
              }}
            </SelectTrigger>
            <SelectPopup>
              {options.map(org => (
                <SelectItem key={org.slug} value={org.slug}>
                  {org.displayName}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
      );
    }
    // The provider has no picker: its organization was captured during the
    // OAuth callback. An empty list is expected here, so do not warn about a
    // connection that is perfectly healthy — but do not offer an input either.
    // No provider in this branch implements SetOrganizationSettings, so
    // submitting one could only ever return "does not support organization
    // configuration". Show what the callback captured, read-only; changing it
    // means reconnecting.
    // A captured organization is still worth showing, but not instead of a
    // refusal: this branch owns the whole trailing cell, so the parent's
    // standalone issue is suppressed here and a broken source would otherwise
    // render as nothing but an organization name.
    case "NOT_APPLICABLE": {
      const issue = connectionIssue(connectionStatus, canReconnect, "NOT_APPLICABLE");
      if (issue != null) {
        return (
          <SourceConnectionIssue
            connectorKey={connectorKey}
            issueKey={issue}
            reconnectUrl={reconnectUrl}
          />
        );
      }

      return <CapturedOrganization sourceKey={source} />;
    }
    case "EMPTY":
      return (
        <ProviderOrganizationsEmpty
          sourceKey={source}
          providerName={providerName}
          onSubmit={onSelect}
        />
      );
    // UNAVAILABLE, plus the impossible case of a node that is not an
    // AccessReviewSource: either way the list could not be read.
    default: {
      const issue = connectionIssue(connectionStatus, canReconnect, "UNAVAILABLE");
      if (issue == null) {
        return null;
      }

      return (
        <SourceConnectionIssue
          connectorKey={connectorKey}
          issueKey={issue}
          reconnectUrl={reconnectUrl}
        />
      );
    }
  }
}
