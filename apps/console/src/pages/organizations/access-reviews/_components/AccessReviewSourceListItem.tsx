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

import { dateTimeFormat } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Select,
  ThirdPartyLogo,
  useConfirm,
} from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessReviewSourceListItem_source$key } from "#/__generated__/core/AccessReviewSourceListItem_source.graphql";
import type { AccessReviewSourceListItemConfigureMutation } from "#/__generated__/core/AccessReviewSourceListItemConfigureMutation.graphql";
import type { AccessReviewSourceListItemDeleteMutation } from "#/__generated__/core/AccessReviewSourceListItemDeleteMutation.graphql";
import type { InlineOrgSelectQuery } from "#/__generated__/core/InlineOrgSelectQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { buildConnectorInitiateURL } from "#/pages/organizations/settings/integrations/_lib/connectorInitiate";

import { connectionIssue } from "../_lib/connectionIssue";

import { InlineOrgSelect, inlineOrgSelectQuery } from "./InlineOrgSelect";
import { SourceConnectionIssue } from "./SourceConnectionIssue";

const fragment = graphql`
  fragment AccessReviewSourceListItem_source on AccessReviewSource {
    id
    name
    connectorId
    connector {
      provider
      protocol
      canReconnect
      displayName
      oauth2Scopes
      ...SourceConnectionIssue_connector
    }
    connectionStatus
    selectedOrganization
    needsConfiguration
    createdAt
    canDelete: permission(action: "access-review:source:delete")
  }
`;

export const deleteAccessReviewSourceMutation = graphql`
  mutation AccessReviewSourceListItemDeleteMutation(
    $input: DeleteAccessReviewSourceInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewSource(input: $input) {
      deletedAccessReviewSourceId @deleteEdge(connections: $connections)
    }
  }
`;

const configureMutation = graphql`
  mutation AccessReviewSourceListItemConfigureMutation(
    $input: ConfigureAccessReviewSourceInput!
  ) {
    configureAccessReviewSource(input: $input) {
      accessReviewSource {
        id
        selectedOrganization
        needsConfiguration
      }
    }
  }
`;

type Props = {
  sourceKey: AccessReviewSourceListItem_source$key;
  connectionId: string;
};

export function AccessReviewSourceListItem({
  sourceKey,
  connectionId,
}: Props) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const confirm = useConfirm();

  const accessSource = useFragment(fragment, sourceKey);

  const [deleteAccessReviewSource]
    = useMutation<AccessReviewSourceListItemDeleteMutation>(
      deleteAccessReviewSourceMutation,
      { errorToast: t("accessReviewSourceRow.errors.delete") },
    );
  const [configure]
    = useMutation<AccessReviewSourceListItemConfigureMutation>(configureMutation, {
      successMessage: t("accessReviewSourceRow.messages.organizationUpdated"),
      errorToast: t("accessReviewSourceRow.errors.configure"),
    });
  const [orgsQueryRef, loadOrgsQuery]
    = useQueryLoader<InlineOrgSelectQuery>(inlineOrgSelectQuery);

  const connector = accessSource.connector;
  const showOrgSelector = connector != null
    && accessSource.connectionStatus !== "NOT_APPLICABLE"
    && Boolean(accessSource.needsConfiguration || accessSource.selectedOrganization);

  useEffect(() => {
    if (!showOrgSelector) {
      return;
    }

    loadOrgsQuery(
      { accessReviewSourceId: accessSource.id },
      { fetchPolicy: "store-or-network" },
    );
  }, [showOrgSelector, loadOrgsQuery, accessSource.id]);

  const readyOrgsQueryRef = orgsQueryRef != null
    && orgsQueryRef.variables.accessReviewSourceId === accessSource.id
    ? orgsQueryRef
    : null;

  const handleDelete = () => {
    confirm(
      () => {
        void deleteAccessReviewSource({
          variables: {
            input: { accessReviewSourceId: accessSource.id },
            connections: [connectionId],
          },
        }).catch(() => undefined);
      },
      {
        message: t("accessReviewSourceRow.deleteConfirmation", {
          name: accessSource.name,
        }),
      },
    );
  };

  const handleOrgChange = (slug: string) => {
    void configure({
      variables: {
        input: {
          accessReviewSourceId: accessSource.id,
          organizationSlug: slug,
        },
      },
    }).catch(() => undefined);
  };

  if (connector == null || accessSource.connectionStatus === "NOT_APPLICABLE") {
    return (
      <Card variant="soft" size={2} className="flex h-full min-w-0 flex-col gap-3">
        <div className="flex items-center gap-4">
          <div className="flex min-w-0 flex-1 flex-col gap-0.5">
            <Heading level={3} size={3} weight="medium" highContrast className="min-w-0 truncate">
              {accessSource.name}
            </Heading>
            <Text size={1} color="faint">
              <time dateTime={accessSource.createdAt}>
                {dateTimeFormat(i18n.language, accessSource.createdAt)}
              </time>
            </Text>
          </div>
        </div>
        {accessSource.canDelete && (
          <div className="flex flex-wrap items-center justify-end gap-2">
            <ActionDropdown>
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  handleDelete();
                }}
              >
                {t("accessReviewSourceRow.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          </div>
        )}
      </Card>
    );
  }

  const connectionStatus = accessSource.connectionStatus;
  const canReconnect = connector.canReconnect;
  const reconnectUrl = accessSource.connectorId == null
    ? null
    : buildConnectorInitiateURL(
        organizationId,
        connector.provider,
        connector.protocol,
        {
          connectorId: accessSource.connectorId,
          oauth2Scopes: connector.oauth2Scopes,
        },
      );
  // Stated as what a healthy source is, not as a list of the ways it can
  // break: a status added to the enum after this bundle shipped would match
  // no branch of such a list and silently render nothing at all for a source
  // that is in fact broken. An unrecognised one is treated as DISCONNECTED is.
  const hasConnectionIssue = connectionStatus !== "CONNECTED";
  const showStandaloneIssue = hasConnectionIssue && !showOrgSelector;
  const standaloneIssue = connectionIssue(
    connectionStatus,
    canReconnect,
    "NOT_APPLICABLE",
  );

  return (
    <Card variant="soft" size={2} className="flex h-full min-w-0 flex-col gap-3">
      <div className="flex items-center gap-4">
        <ThirdPartyLogo
          thirdParty={connector.provider}
          className="size-8 shrink-0"
        />
        <div className="flex min-w-0 flex-1 flex-col gap-0.5">
          <Heading level={3} size={3} weight="medium" highContrast className="min-w-0 truncate">
            {accessSource.name}
          </Heading>
          <Text size={1} color="faint">
            <time dateTime={accessSource.createdAt}>
              {dateTimeFormat(i18n.language, accessSource.createdAt)}
            </time>
          </Text>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-end gap-2">
        {showOrgSelector && (
          readyOrgsQueryRef == null
            ? <OrgSelectPending />
            : (
                <Suspense fallback={<OrgSelectPending />}>
                  <InlineOrgSelect
                    queryRef={readyOrgsQueryRef}
                    connectionStatus={connectionStatus}
                    canReconnect={canReconnect}
                    reconnectUrl={canReconnect ? reconnectUrl : null}
                    connectorKey={connector}
                    providerName={connector.displayName}
                    onSelect={handleOrgChange}
                  />
                </Suspense>
              )
        )}
        {showStandaloneIssue && standaloneIssue != null && (
          <SourceConnectionIssue
            connectorKey={connector}
            issueKey={standaloneIssue}
            reconnectUrl={canReconnect ? reconnectUrl : null}
          />
        )}
        {accessSource.canDelete && (
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onSelect={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleDelete();
              }}
            >
              {t("accessReviewSourceRow.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>
    </Card>
  );
}

function OrgSelectPending() {
  const { t } = useTranslation();

  return (
    <Select
      variant="editor"
      disabled
      placeholder={t("accessReviewSourceRow.loading")}
    />
  );
}
